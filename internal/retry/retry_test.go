package retry_test

import (
	"errors"
	"fmt"

	"github.com/jackc/pgconn"

	"github.com/FollowLille/metrics/internal/retry"
)

// Пример использования функции retry.Retry для повторения операции с настройкой задержек.
func ExampleRetry() {
	operation := func() error {
		fmt.Println("Попытка выполнения операции")
		// Возвращаем ошибку для демонстрации повторов.
		return errors.New("временная ошибка")
	}

	// Выполняем операцию с повторами.
	err := retry.Retry(operation)
	if err != nil {
		fmt.Println("Операция завершилась с ошибкой:", err)
	} else {
		fmt.Println("Операция успешно выполнена")
	}

	// Output:
	// Попытка выполнения операции
	// Попытка выполнения операции
	// Попытка выполнения операции
	// Операция завершилась с ошибкой: временная ошибка
}

// Пример использования функции retry.IsRetriablePostgresError для проверки ошибок PostgreSQL.
func ExampleIsRetriablePostgresError() {
	// Пример ошибки, которая повторяется.
	retriableErr := &pgconn.PgError{Code: "40001"} // Serialization failure.

	// Пример ошибки, которая не повторяется.
	nonRetriableErr := &pgconn.PgError{Code: "23505"} // Unique violation.

	// Проверяем ошибки.
	fmt.Println("Ошибка повторяемая:", retry.IsRetriablePostgresError(retriableErr))
	fmt.Println("Ошибка не повторяемая:", retry.IsRetriablePostgresError(nonRetriableErr))

	// Output:
	// Ошибка повторяемая: true
	// Ошибка не повторяемая: false
}
