package flags

import (
	"flag"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
)

var (
	FlagRunAddr        string
	FlagReportInterval int
	FlagPollInterval   int
	FlagHashKey        string
	UseHashKey         bool
)

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func ParseFlags() {

	//defaultHashKey := "0123456789ABCDEF"
	defaultHashKey := ""

	flag.StringVar(&FlagRunAddr, "a", "localhost:8080", "addr to run on")
	flag.IntVar(&FlagReportInterval, "r", 10, "reportInterval")
	flag.IntVar(&FlagPollInterval, "p", 2, "pollInterval")
	flag.StringVar(&FlagHashKey, "k", defaultHashKey, "hashKey")
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

	if envHashKey := os.Getenv("KEY"); envHashKey != "" {
		FlagHashKey = envHashKey
	}

	//UseHashKey = true
	if !isFlagPassed(FlagHashKey) && os.Getenv("KEY") == "" {
		UseHashKey = false
	} else {
		UseHashKey = true
	}
	log.Info().
		Str("UseHashKey", strconv.FormatBool(UseHashKey)).
		Str("FlagHashKey", FlagHashKey).
		Msg("UseHashKey")

}
