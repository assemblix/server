package main

import (
	"embed"
	"flag"
	"fmt"
	"net/http"
)

//go:embed pages/*
var pages embed.FS

//go:embed cdn/*
var cdn embed.FS

var exit = make(chan error, 1)

var noshell bool

func main() {
	flag.BoolVar(&noshell, "no-shell", false, "disables the cli")
	flag.Parse()

	defer db.Close()

	http.HandleFunc("/", root)
	http.HandleFunc("/register", register)
	http.HandleFunc("/home", home)
	http.HandleFunc("/settings", settings)
	http.HandleFunc("/login", login)

	http.HandleFunc("/api/v1/{endpoint...}", apiv1)
	http.Handle("/cdn/", http.FileServer(http.FS(cdn)))

	if !noshell {
		go func() {
			exit <- cli()
		}()
	}
	go func() {
		if err := http.ListenAndServe(port, nil); err != nil {
			logError(err)
			fmt.Println(err)
			exit <- err
		}
	}()

	logInfo(fmt.Errorf("server started"))

	err := <-exit
	if err != nil {
		fmt.Println(err)
	}
}
