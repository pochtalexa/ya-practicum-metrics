package flags

import (
	"flag"
	"os"
	"strconv"
)

var (
	FlagRunAddr        string
	FlagReportInterval int
	FlagPollInterval   int
)

// ParseFlags -a 8090 -r 2 -p 1
func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", "localhost:8080", "addr to run on")
	flag.IntVar(&FlagReportInterval, "r", 10, "reportInterval")
	flag.IntVar(&FlagPollInterval, "p", 2, "pollInterval")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		FlagRunAddr = envRunAddr
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		FlagReportInterval, _ = strconv.Atoi(envReportInterval)
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		FlagPollInterval, _ = strconv.Atoi(envPollInterval)
	}
}
