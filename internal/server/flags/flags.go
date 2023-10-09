package flags

import (
	"flag"
	"fmt"
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
	FlagDbConn        string
	err               error
)

func ParseFlags() {
	defaultFileStorePath := "/tmp/metrics-db.json"
	if opSyst := runtime.GOOS; strings.Contains(opSyst, "windows") {
		defaultFileStorePath = "c:/tmp/metrics-db.json"
	}

	defaultDbConn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		`localhost`, `5432`, `praktikum`, `praktikum`, `praktikum`)

	flag.StringVar(&FlagRunAddr, "a", "localhost:8080", "addr to run on")
	flag.IntVar(&FlagStoreInterval, "i", 300, "save to file interval (sec)")
	flag.StringVar(&FlagFileStorePath, "f", defaultFileStorePath, "file to save")
	flag.BoolVar(&FlagRestore, "r", true, "load metrics on start from file")
	flag.StringVar(&FlagDbConn, "d", defaultDbConn, "db conn string")
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

	if envVar := os.Getenv("DATABASE_DSN"); envVar != "" {
		FlagDbConn = envVar
		if err != nil {
			panic(err)
		}
	}

}
