package retry

import (
	"errors"
	"github.com/FollowLille/metrics/internal/config"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

var (
	ConnectionError           = errors.New("connection error")
	ServerError               = errors.New("server error")
	NonRetriableError         = errors.New("not retriable error")
	NonRetriablePostgresError = errors.New("non retriable postgres error")
	RetriablePostgresError    = errors.New("retriable postgres error")
)

func Retry(operation func() error) error {

	var err error
	for _, delay := range config.DatabaseRetryDelays {
		err = operation()
		if err == nil {
			return nil
		}
		if err == NonRetriableError {
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
