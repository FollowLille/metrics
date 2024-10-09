package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/FollowLille/metrics/internal/retry"
	"log"
	"time"

	"github.com/FollowLille/metrics/internal/logger"
	"github.com/FollowLille/metrics/internal/storage"
	"go.uber.org/zap"
)

var DB *sql.DB

func InitDB(connStr string) {
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		logger.Log.Fatal("can't open database", zap.Error(err))
		return
	}

	err = DB.Ping()
	if err != nil {
		logger.Log.Fatal("can't ping database", zap.Error(err))
		return
	}

	log.Println("Successfully connected to the database")
}

func PrepareDB() {
	_, err := DB.Exec("CREATE SCHEMA IF NOT EXISTS metrics")
	if err != nil {
		logger.Log.Error("can't create schema", zap.Error(err))
		return
	}

	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS metrics.metrics (load_id int not null, metric_type text not null, metric_name text not null, gauge_value double precision, counter_value int)")
	if err != nil {
		logger.Log.Error("can't create table", zap.Error(err))
	}
}

func SaveMetricsToDatabase(db *sql.DB, s *storage.MemStorage) error {
	gauge := s.GetAllGauges()
	counter := s.GetAllCounters()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("can't ping database: %s", err)
	}

	var maxID int64
	err := QueryRowWithRetry(ctx, db, "SELECT COALESCE(MAX(load_id), 0) FROM metrics.metrics", &maxID)
	if err != nil {
		return fmt.Errorf("can't get max id: %s", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("can't begin transaction", zap.Error(err))
		return fmt.Errorf("can't begin transaction: %s", err)
	}
	defer tx.Rollback()

	for id, value := range gauge {
		query := "INSERT INTO metrics.metrics (load_id, metric_name, metric_type, gauge_value) VALUES ($1, $2, $3, $4)"
		err = ExecQueryWithRetry(ctx, tx, query, maxID+1, id, "gauge", value)
		if err != nil {
			logger.Log.Error("can't insert gauge", zap.Error(err))
			return fmt.Errorf("can't insert gauge: %s", err)
		}
	}

	for id, value := range counter {
		query := "INSERT INTO metrics.metrics (load_id, metric_name, metric_type, counter_value) VALUES ($1, $2, $3, $4)"
		err = ExecQueryWithRetry(ctx, tx, query, maxID+1, id, "counter", value)
		if err != nil {
			logger.Log.Error("can't insert counter", zap.Error(err))
			return fmt.Errorf("can't insert counter: %s", err)
		}
	}

	if err = tx.Commit(); err != nil {
		logger.Log.Error("can't commit transaction", zap.Error(err))
		return fmt.Errorf("can't commit transaction: %s", err)
	}

	logger.Log.Info("metrics successfully saved to the database")
	return nil
}

func LoadMetricsFromDatabase(str *storage.MemStorage, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("can't ping database: %s", err)
	}

	var maxID int64
	err := QueryRowWithRetry(ctx, db, "SELECT COALESCE(MAX(load_id), 0) FROM metrics.metrics", &maxID)
	if err != nil {
		return fmt.Errorf("can't get max id: %s", err)
	}

	var metricName string
	var gaugeValue float64
	var counterValue int64

	gaugeRows, err := QueryRowsWithRetry(ctx, db, "SELECT metric_name, gauge_value FROM metrics.metrics WHERE load_id = $1 and metric_type = 'gauge'", maxID)
	if err != nil {
		logger.Log.Error("can't get gauge", zap.Error(err))
		return fmt.Errorf("can't get gauge: %s", err)
	}
	defer gaugeRows.Close()

	for gaugeRows.Next() {
		err = gaugeRows.Scan(&metricName, &gaugeValue)
		if err != nil {
			logger.Log.Error("can't scan gauge", zap.Error(err))
			return fmt.Errorf("can't scan gauge: %s", err)
		}
		str.UpdateGauge(metricName, gaugeValue)
	}

	if err = gaugeRows.Err(); err != nil {
		logger.Log.Error("can't get gauge", zap.Error(err))
		return fmt.Errorf("can't get gauge: %s", err)
	}

	counterRows, err := QueryRowsWithRetry(ctx, db, "SELECT metric_name, counter_value FROM metrics.metrics WHERE load_id = $1 and metric_type = 'counter'", maxID)
	if err != nil {
		logger.Log.Error("can't get counter", zap.Error(err))
		return fmt.Errorf("can't get counter: %s", err)
	}
	defer counterRows.Close()

	for counterRows.Next() {
		err = counterRows.Scan(&metricName, &counterValue)
		if err != nil {
			logger.Log.Error("can't scan counter", zap.Error(err))
			return fmt.Errorf("can't scan counter: %s", err)
		}
		str.UpdateCounter(metricName, counterValue)
	}

	if err = counterRows.Err(); err != nil {
		logger.Log.Error("can't get counter", zap.Error(err))
		return fmt.Errorf("can't get counter: %s", err)
	}

	return nil
}

// ExecContexter interface нужен чтобы функции записи\чтения умели работать как с sql.DB так и с sql.Tx
type ExecContexter interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func ExecQueryWithRetry(ctx context.Context, exec ExecContexter, query string, agrs ...interface{}) error {
	err := retry.Retry(func() error {
		_, execErr := exec.ExecContext(ctx, query, agrs...)
		if execErr != nil {
			if retry.IsRetriablePostgresError(execErr) {
				logger.Log.Error("retriable postgres error", zap.Error(execErr))
				return retry.RetriablePostgresError
			}
			logger.Log.Error("non retriable postgres error", zap.Error(execErr))
			return retry.NonRetriablePostgresError
		}
		return nil
	})

	if err != nil {
		logger.Log.Error("can't execute query", zap.Error(err))
		return err
	}

	logger.Log.Info("query successfully executed", zap.String("query", query))
	return nil
}

func QueryRowWithRetry(ctx context.Context, db *sql.DB, query string, dest ...interface{}) error {
	err := retry.Retry(func() error {
		row := db.QueryRowContext(ctx, query)
		if scanErr := row.Scan(dest...); scanErr != nil {
			if retry.IsRetriablePostgresError(scanErr) {
				logger.Log.Error("retriable postgres error", zap.Error(scanErr))
				return retry.RetriablePostgresError
			}
			logger.Log.Error("non retriable postgres error", zap.Error(scanErr))
			return retry.NonRetriablePostgresError
		}
		return nil
	})

	if err != nil {
		logger.Log.Error("can't execute query", zap.Error(err))
		return err
	}

	logger.Log.Info("query successfully executed", zap.String("query", query))
	return nil
}

func QueryRowsWithRetry(ctx context.Context, db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	var rows *sql.Rows
	var err error

	err = retry.Retry(func() error {
		rows, err = db.QueryContext(ctx, query, args...)
		if err != nil {
			if retry.IsRetriablePostgresError(err) {
				logger.Log.Error("retriable postgres error", zap.Error(err))
				return retry.RetriablePostgresError
			}
			logger.Log.Error("non retriable postgres error", zap.Error(err))
			return retry.NonRetriablePostgresError
		}
		return nil
	})

	if err != nil {
		logger.Log.Error("can't execute query", zap.Error(err))
		return nil, err
	}

	logger.Log.Info("query successfully executed", zap.String("query", query))
	return rows, nil
}
