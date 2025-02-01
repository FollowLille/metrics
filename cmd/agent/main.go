// Package main отвечает за инициализацию и запуск агента
// Он включает в себя функции для парсинга командных флагов и переменных окружения,
// а также настройку логгирования.
package main

import (
	"crypto/rsa"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/FollowLille/metrics/internal/agent"
	"github.com/FollowLille/metrics/internal/crypto"
	"github.com/FollowLille/metrics/internal/logger"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
	publicKey    *rsa.PublicKey
)

func main() {
	PrintBuildFlag(buildVersion, buildDate, buildCommit)
	err := parseFlags()
	if err != nil {
		fmt.Printf("invalid flags: %s", err)
		return
	}

	a := Init(flagAddress, flagCryptoKeyPath)
	logger.Initialize("info")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go a.Run()

	sig := <-sigChan
	logger.Log.Info("received signal", zap.String("signal", sig.String()))
	a.Shutdown()

	time.Sleep(5 * time.Second)
	logger.Log.Info("agent shutdown")
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
func Init(flags string, flagCryptoKeyPath string) *agent.Agent {
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

	if flagCryptoKeyPath != "" {
		publicKey, err = crypto.LoadPublicKey(flagCryptoKeyPath)
		if err != nil {
			logger.Log.Fatal("failed to load public key", zap.Error(err))
		}
		a.PublicKey = publicKey
	}
	a.ServerAddress = serverAddress
	a.ServerPort = serverPort
	a.HashKey = flagHashKey
	a.PollInterval = time.Duration(flagPollInterval) * time.Second
	a.ReportSendInterval = time.Duration(flagReportInterval) * time.Second
	a.RateLimit = flagRateLimit

	return a
}

// PrintBuildFlag выводит информацию о версии сборки, дате сборки и коммите.
// Если переменные пусты, выводит "N/A".
func PrintBuildFlag(buildVersion, buildDate, buildCommit string) {
	buildVersion = IfFlagEmpty(buildVersion, "N/A")
	buildDate = IfFlagEmpty(buildDate, "N/A")
	buildCommit = IfFlagEmpty(buildCommit, "N/A")

	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)
}

// IfFlagEmptyвозвращает значение `flag`, если оно не пустое. В противном случае возвращает `alternative`.
func IfFlagEmpty(flag, alternative string) string {
	if flag == "" {
		return alternative
	}
	return flag
}
