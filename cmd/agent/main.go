// Package main отвечает за инициализацию и запуск агента
// Он включает в себя функции для парсинга командных флагов и переменных окружения,
// а также настройку логгирования.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/FollowLille/metrics/internal/agent"
	"github.com/FollowLille/metrics/internal/logger"
)

func main() {
	err := parseFlags()
	if err != nil {
		fmt.Printf("invalid flags: %s", err)
		os.Exit(1)
	}
	a := Init(flagAddress)
	logger.Initialize("info")
	a.Run()
}

// Init инициализирует агента
// Принимает флаг адреса и пытается его обработать как хост:порт
// Если адрес некорректный, то выходит с ошибкой
// Если адрес корректный, то инициализирует агента
//
// Параметры:
//   - flags - адрес для прослушивания
//
// Возвращаемое значение:
//   - agent.Agent - инициализированный агент
func Init(flags string) *agent.Agent {
	splitedAddress := strings.Split(flags, ":")
	if len(splitedAddress) != 2 {
		fmt.Printf("invalid address %s, expected host:port", flags)
		os.Exit(1)
	}
	serverAddress := splitedAddress[0]
	serverPort, err := strconv.ParseInt(splitedAddress[1], 10, 64)
	if err != nil {
		fmt.Printf("invalid port: %d", serverPort)
		os.Exit(1)
	}

	a := agent.NewAgent()
	a.ServerAddress = serverAddress
	a.ServerPort = serverPort
	a.HashKey = flagHashKey
	a.PollInterval = time.Duration(flagPollInterval) * time.Second
	a.ReportSendInterval = time.Duration(flagReportInterval) * time.Second
	a.RateLimit = flagRateLimit

	return a
}
