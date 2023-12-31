package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/jhelison/go-torrent/cmd"
)

func main() {
	// Debugger
	go func() {
		// We don't really care for the error
		//nolint:errcheck
		http.ListenAndServe("localhost:8080", nil)
	}()

	// Run the cobra CMD
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
