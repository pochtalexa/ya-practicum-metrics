package flags

import (
	"flag"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var (
	FlagRunAddr       string
	FlagStoreInterval int
	FlagFileStorePath string
	FlagRestore       bool
	err               error
)

func ParseFlags() {
	defaultFileStorePath := "/tmp/metrics-db.json"
	if opSyst := runtime.GOOS; strings.Contains(opSyst, "windows") {
		defaultFileStorePath = "c:/tmp/metrics-db.json"
	}

	flag.StringVar(&FlagRunAddr, "a", "localhost:8080", "addr to run on")
	flag.IntVar(&FlagStoreInterval, "i", 300, "save to file interval (sec)")
	flag.StringVar(&FlagFileStorePath, "f", defaultFileStorePath, "file to save")
	flag.BoolVar(&FlagRestore, "r", true, "load metrics on start from file")
	flag.Parse()

	if envVar := os.Getenv("ADDRESS"); envVar != "" {
		FlagRunAddr = envVar
	}

	if envVar := os.Getenv("STORE_INTERVAL"); envVar != "" {
		FlagStoreInterval, err = strconv.Atoi(envVar)
		if err != nil {
			panic(err)
		}
	}

	if envVar := os.Getenv("FILE_STORAGE_PATH"); envVar != "" {
		FlagFileStorePath = envVar
	}

	if envVar := os.Getenv("RESTORE"); envVar != "" {
		FlagRestore, err = strconv.ParseBool(envVar)
		if err != nil {
			panic(err)
		}
	}
}
