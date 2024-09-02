package main

import (
	"github.com/FollowLille/metrics/internal/agent"
)

func main() {
	a := agent.NewAgent()
	a.Run()
}
