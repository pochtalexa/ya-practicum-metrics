package flags

import "flag"

var (
	FlagRunAddr        string
	FlagReportInterval int
	FlagPollInterval   int
)

// ParseFlags -a 8090 -r 2 -p 1
func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "addr to run on")
	flag.IntVar(&FlagReportInterval, "r", 10, "reportInterval")
	flag.IntVar(&FlagPollInterval, "p", 2, "pollInterval")
	flag.Parse()
}
