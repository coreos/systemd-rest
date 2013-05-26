/*
*  Copyright 2013 CoreOS, Inc
*
*  Licensed under the Apache License, Version 2.0 (the "License");
*  you may not use this file except in compliance with the License.
*  You may obtain a copy of the License at
*
*      http://www.apache.org/licenses/LICENSE-2.0
*
*  Unless required by applicable law or agreed to in writing, software
*  distributed under the License is distributed on an "AS IS" BASIS,
*  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*  See the License for the specific language governing permissions and
*  limitations under the License.
 */

package main

import (
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Options struct {
	Dir  string
	Port string
}

var options = Options{}

func init() {
	const (
		defaultDir  = "/"
		defaultPort = "8080"
	)

	flag.StringVar(&options.Dir, "D", defaultDir, "Directory prefix (default /)")
	flag.StringVar(&options.Port, "p", defaultPort, "Port to bind to")
}

const StateDir = "/var/lib/systemd-rest/"

func main() {
	flag.Parse()

	r := mux.NewRouter()

	setupUnits(r.PathPrefix("/units").Subrouter(), options)
	setupDocker(r.PathPrefix("/docker").Subrouter(), options)

	http.Handle("/", r)
	err := http.ListenAndServe(":"+options.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
