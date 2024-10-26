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

func Init(flags string) agent.Agent {
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

	a := agent.Agent{}
	a.ServerAddress = serverAddress
	a.ServerPort = serverPort
	a.HashKey = flagHashKey
	a.PollInterval = time.Duration(flagPollInterval) * time.Second
	a.ReportSendInterval = time.Duration(flagReportInterval) * time.Second
	return a
}
