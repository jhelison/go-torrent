package main

import (
	"go-torrent/cmd"
	"net/http"
	_ "net/http/pprof"
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
