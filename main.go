package main

import (
	"encoding/json"
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

func listHandler(w http.ResponseWriter, r *http.Request) {
	var (
		out interface{}
	)

	s := new(systemd.Systemd1)
	err := s.Connect()
	if err != nil {
		// TODO: Return 40* code
		fmt.Fprint(w, err)
	}

	out, err = s.ListUnits()

	if err != nil {
		// TODO: Return 40* code
		fmt.Fprint(w, err)
	}

	outJson, _ := json.Marshal(out)
	fmt.Println("%s\n", outJson)
	fmt.Fprint(w, "%s\n", outJson)
}

func unitHandler(w http.ResponseWriter, r *http.Request) {
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
		out, err = s.StopUnit(vars["unit"], vars["mode"])
	}

	if err != nil {
		w.WriteHeader(404)
	}

	outJson, _ := json.Marshal(out)
	fmt.Fprintf(w, "%s\n", outJson)
}

func main() {
	op := Options{Path: "./", Port: "8080"}

	r := mux.NewRouter()
	r.HandleFunc("/units", listHandler)
	r.HandleFunc("/units/{unit}/{method}/{mode}", unitHandler)

	http.Handle("/", r)
	err := http.ListenAndServe(":" + op.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

