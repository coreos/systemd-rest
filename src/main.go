package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/philips/go-systemd"
)

type Options struct {
	Path string
	Port string
}

func handler(w http.ResponseWriter, r *http.Request) {
	s := new(systemd.Systemd1)
	err := s.Connect()
	if err != nil {
		// TODO: Return 40* code
		fmt.Fprint(w, err)
	}

	out, err := s.StartUnit("acpid.service", "fail")
	if err != nil {
		// TODO: Return 40* code
		fmt.Fprint(w, err)
	}

	fmt.Fprint(w, out)
}

func main() {
	op := Options{Path: "./", Port: "8080"}

	r := mux.NewRouter()
	r.HandleFunc("/", handler).
		Methods("GET")

	http.Handle("/", r)
	err := http.ListenAndServe(":" + op.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

