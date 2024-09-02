package main

import (
	"fmt"
	"net/http"

	"github.com/FollowLille/metrics/internal/handler"
	"github.com/FollowLille/metrics/internal/server"
	"github.com/FollowLille/metrics/internal/storage"
)

func main() {
	metricsStorage := storage.NewMemStorage()
	mux := http.NewServeMux()

	// Просто для проверки корректного запуска
	mux.HandleFunc("/", handler.HomeHandler)

	// Обработчик обновлений
	mux.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		handler.UpdateHandler(w, r, metricsStorage)
	})

	// Запуск HTTP-сервера
	s := server.NewServer()
	addr := fmt.Sprintf("%s:%d", s.Address, s.Port)
	err := http.ListenAndServe(addr, mux)

	if err != nil {
		panic(err)
	}
}
