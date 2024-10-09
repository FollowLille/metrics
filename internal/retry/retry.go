package retry

import (
	"errors"
	"github.com/FollowLille/metrics/internal/config"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

var (
	ErrorConnection           = errors.New("connection error")
	ErrorServer               = errors.New("server error")
	ErrorNonRetriable         = errors.New("not retriable error")
	ErrorNonRetriablePostgres = errors.New("non retriable postgres error")
	ErrorRetriablePostgres    = errors.New("retriable postgres error")
)

func Retry(operation func() error) error {

	var err error
	for _, delay := range config.DatabaseRetryDelays {
		err = operation()
		if err == nil {
			return nil
		}
		if err == ErrorNonRetriable {
			return err
		}
		time.Sleep(delay)
	}

	return err
}

func IsRetriablePostgresError(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case pgerrcode.ConnectionException,
			pgerrcode.ConnectionFailure,
			pgerrcode.AdminShutdown,
			pgerrcode.SerializationFailure,
			pgerrcode.DeadlockDetected:
			return true
		}
	}
	return false
}
