package flags

import (
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type StoragePoint struct {
	Memory   bool
	File     bool
	DataBase bool
}

var (
	FlagRunAddr       string
	FlagStoreInterval int
	FlagFileStorePath string
	FlagRestore       bool
	FlagDBConn        string
	StorePoint        StoragePoint
	err               error
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
	defaultFileStorePath := "/tmp/metrics-db.json"
	if opSyst := runtime.GOOS; strings.Contains(opSyst, "windows") {
		defaultFileStorePath = "c:/tmp/metrics-db.json"
	}

	defaultDBConn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		`localhost`, `5432`, `praktikum`, `praktikum`, `praktikum`)

	flag.StringVar(&FlagRunAddr, "a", ":8080", "addr to run on")
	flag.IntVar(&FlagStoreInterval, "i", 300, "save to file interval (sec)")
	flag.StringVar(&FlagFileStorePath, "f", defaultFileStorePath, "file to save")
	flag.BoolVar(&FlagRestore, "r", true, "load metrics on start from file")
	flag.StringVar(&FlagDBConn, "d", defaultDBConn, "db conn string")
	flag.Parse()

	log.Info().Str("isFlagPassed: FlagRunAddr", strconv.FormatBool(isFlagPassed(FlagRunAddr))).Msg("")
	log.Info().Str("FlagRunAddr", FlagRunAddr).Msg("")
	if envVar := os.Getenv("ADDRESS"); envVar != "" {
		log.Info().Str("ADDRESS", envVar).Msg("")
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

	log.Info().Str("isFlagPassed: FlagDBConn", strconv.FormatBool(isFlagPassed(FlagDBConn))).Msg("")
	log.Info().Str("FlagDBConn", FlagDBConn).Msg("")
	log.Info().Str("Getenv DATABASE_DSN", os.Getenv("DATABASE_DSN")).Msg("")
	if envVar := os.Getenv("DATABASE_DSN"); envVar != "" {
		log.Info().Str("DATABASE_DSN", envVar).Msg("")
		FlagDBConn = envVar
		if err != nil {
			panic(err)
		}
	}

	if isFlagPassed(FlagDBConn) || os.Getenv("DATABASE_DSN") != "" {
		StorePoint.DataBase = true
	} else if isFlagPassed(FlagFileStorePath) || FlagFileStorePath != "" {
		StorePoint.File = true
	} else {
		StorePoint.Memory = true
	}
}
