package main

import (
	"go-torrent/cmd"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	// Debugger
	go func() {
		http.ListenAndServe("localhost:8080", nil)
	}()

	// Run the cobra CMD
	cmd.Execute()
}
