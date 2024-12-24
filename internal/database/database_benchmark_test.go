package database

import (
	"database/sql"
	"testing"

	"github.com/FollowLille/metrics/internal/storage"
)

func BenchmarkSaveMetricsToDatabase(b *testing.B) {
	// Настраиваем тестовую среду
	db, _ := sql.Open("postgres", "postgres://praktikum_go:userpassword@localhost:5432/metrics?sslmode=disable") // Замените на ваш DSN
	defer db.Close()

	memStorage := storage.NewMemStorage()
	memStorage.UpdateGauge("test_gauge", 123.456)
	memStorage.UpdateCounter("test_counter", 789)

	// Benchmark loop
	for i := 0; i < b.N; i++ {
		err := SaveMetricsToDatabase(db, memStorage)
		if err != nil {
			b.Errorf("error in SaveMetricsToDatabase: %v", err)
		}
	}
}

func BenchmarkLoadMetricsFromDatabase(b *testing.B) {
	// Настраиваем тестовую среду
	db, _ := sql.Open("postgres", "postgres://praktikum_go:userpassword@localhost:5432/metrics?sslmode=disable") // Замените на ваш DSN
	defer db.Close()

	memStorage := storage.NewMemStorage()

	// Benchmark loop
	for i := 0; i < b.N; i++ {
		err := LoadMetricsFromDatabase(memStorage, db)
		if err != nil {
			b.Errorf("error in LoadMetricsFromDatabase: %v", err)
		}
	}
}
