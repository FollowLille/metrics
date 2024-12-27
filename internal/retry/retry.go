// Package retry реализует повторные попытки выполнения операции
package retry

import (
	"errors"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	"github.com/FollowLille/metrics/internal/config"
)

var (
	ErrorConnection           = errors.New("connection error")             // ошибка соединения
	ErrorServer               = errors.New("server error")                 // ошибка сервера
	ErrorNonRetriable         = errors.New("not retriable error")          // не повторяемая ошибка
	ErrorNonRetriablePostgres = errors.New("non retriable postgres error") // не повторяемая ошибка postgres
	ErrorRetriablePostgres    = errors.New("retriable postgres error")     // повторяемая ошибка postgres
)

// Retry повторяет выполнение операции до тех пор, пока она не завершится без ошибок
// Принимает функцию, которая выполняет операцию
// Возвращает ошибку, если она возникнет
//
// Параметры:
//   - operation - функция, которая выполняет операцию
//
// Возвращаемое значение:
//   - error - ошибка
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

// IsRetriablePostgresError проверяет, является ли ошибка повторяемой postgres
// Принимает ошибку
// Возвращает булевое значение
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
