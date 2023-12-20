package main

import (
	"embed"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed pages/*
var pages embed.FS

func main() {
	http.HandleFunc("/", server)
	http.HandleFunc("/api/v1/", apiv1)

	go func() {
		cli()
		exit()
	}()

	go func() {
		if err := http.ListenAndServe(port, nil); err != nil {
			fmt.Println(err)
			exit()
		}
	}()

	select {
	case <-exitChan:
	}

	db.Close()
}

var exitChan = make(chan struct{})

func exit() {
	exitChan <- struct{}{}
}
