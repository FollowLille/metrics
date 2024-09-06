package main

import (
	"fmt"
	"github.com/FollowLille/metrics/internal/agent"
	"os"
	"time"
)

func main() {
	a := agent.NewAgent()
	err := parseFlags()
	a.PollInterval = time.Duration(flagPollInterval) * time.Second
	a.ReportSendInterval = time.Duration(flagReportInterval) * time.Second
	a.ServerPort = flagPort
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	a.Run()
}
