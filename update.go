/*
*  Copyright 2013 CoreOS, Inc
*  Copyright 2013 Docker Authors
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
	"github.com/gorilla/mux"
	"net/http"
	"os/exec"
	"io"
	"fmt"
)

// TODO(bp): Use DBUS endpoints and make this JSON!
func updateHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("update_engine_client", "-update", "-omaha_url=http://update.core-os.net")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, err)
		return
	}
	if err := cmd.Start(); err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, err)
		return
	}
	io.Copy(w, stdout)
	if err := cmd.Wait(); err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, err)
		return
	}
	return
}

func setupUpdate(r *mux.Router, o Options) {
	r.HandleFunc("", updateHandler)
	r.HandleFunc("/", updateHandler)

	return
}
