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
	var (
		out interface{}
	)

	vars := mux.Vars(r)
	s := new(systemd.Systemd1)
	err := s.Connect()
	if err != nil {
		// TODO: Return 40* code
		fmt.Fprint(w, err)
	}

	switch vars["method"] {
	case "start":
		out, err = s.StartUnit(vars["unit"], vars["mode"])
	case "stop":
		out, err = s.StartUnit(vars["unit"], vars["mode"])
	}

	if err != nil {
		// TODO: Return 40* code
		fmt.Fprint(w, err)
	}

	fmt.Fprint(w, out)
}

func main() {
	op := Options{Path: "./", Port: "8080"}

	r := mux.NewRouter()
	r.HandleFunc("/units/{unit}/{method}/{mode}", handler).
		Methods("POST")

	http.Handle("/", r)
	err := http.ListenAndServe(":" + op.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

