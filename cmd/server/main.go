package main

import (
	"fmt"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/flags"
	"net/http"

	"github.com/pochtalexa/ya-practicum-metrics/internal/server/handlers"
)

func run() error {
	fmt.Println("Running server on", flags.FlagRunAddr)
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/update/", handlers.UpdateHandler)
	mux.HandleFunc("/value/", handlers.ValueHandler)
	return http.ListenAndServe(flags.FlagRunAddr, mux)
}

func main() {
	flags.ParseFlags()
	if err := run(); err != nil {
		panic(err)
	}
}
