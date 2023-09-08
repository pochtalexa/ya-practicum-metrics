package main

import (
	"fmt"
	"net/http"	
	// "errors"

	"github.com/pochtalexa/ya-practicum-metrics/internal/server/storage"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/handlers"
)


func run() error {
	mux := http.NewServeMux()	
	mux.HandleFunc("/update/", handlers.UpdateHandler)
    return http.ListenAndServe(":8080", mux)
}

func main() {

	MemStorage := storage.NewMemStore()
	fmt.Println(MemStorage)

	if err := run(); err != nil {
        panic(err)
    }
	
}
