package flags

import (
	"flag"
	"os"
)

var FlagRunAddr string

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", "localhost:8080", "addr to run on")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		FlagRunAddr = envRunAddr
	}
}
