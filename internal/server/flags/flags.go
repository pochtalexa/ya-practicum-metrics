package flags

import "flag"

var FlagRunAddr string

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "addr to run on")
	flag.Parse()
}
