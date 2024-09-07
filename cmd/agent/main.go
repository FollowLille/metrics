package main

import (
	"fmt"
	"github.com/FollowLille/metrics/internal/agent"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	a := agent.NewAgent()
	err := parseFlags()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	a.PollInterval = time.Duration(flagPollInterval) * time.Second
	a.ReportSendInterval = time.Duration(flagReportInterval) * time.Second

	splitedAddress := strings.Split(flagAddress, ":")
	serverAddress := splitedAddress[0]
	serverPort, err := strconv.ParseInt(splitedAddress[1], 10, 64)
	if err != nil {
		fmt.Printf("некорректный адрес: %s", flagAddress)
		os.Exit(1)
	}
	a.ServerAddress = serverAddress
	a.ServerPort = serverPort
	a.Run()
}
