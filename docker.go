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
	"os"
	"path"
	"fmt"
	"log"
	"net/http"
	"github.com/dotcloud/docker"
	"github.com/dotcloud/docker/registry"
	"github.com/gorilla/mux"
)

type Context struct {
	Path string
	Registry *registry.Registry
	Graph *docker.Graph
}

var context Context

func pullImage(c *Context, imgId, registry string, token []string) error {
	history, err := c.Registry.GetRemoteHistory(imgId, registry, token)
	if err != nil {
		return err
	}

	// FIXME: Try to stream the images?
	// FIXME: Launch the getRemoteImage() in goroutines
	for _, id := range history {
		if !c.Graph.Exists(id) {
			log.Printf("Pulling %s metadata\r\n", id)
			imgJson, err := c.Registry.GetRemoteImageJson(id, registry, token)
			if err != nil {
				// FIXME: Keep goging in case of error?
				return err
			}
			img, err := docker.NewImgJson(imgJson)
			if err != nil {
				return fmt.Errorf("Failed to parse json: %s", err)
			}

			// Get the layer
			log.Printf("Pulling %s fs layer\r\n", img.Id)
			layer, _, err := c.Registry.GetRemoteImageLayer(img.Id, registry, token)
			if err != nil {
				return err
			}
			if err := c.Graph.Register(layer, false, img); err != nil {
				return err
			}
		}
	}
	return nil
}


func pullHandler(w http.ResponseWriter, r *http.Request, c *Context) {
	remote := "philips/nova-agent"
	repoData, err := c.Registry.GetRepositoryData(remote)
	if err != nil {
		log.Fatal(err)
	}

	tagsList, err := c.Registry.GetRemoteTags(repoData.Endpoints, remote, repoData.Tokens)
	if err != nil {
		log.Fatal(err)
	}

	for tag, id := range tagsList {
		repoData.ImgList[id].Tag = tag
	}

	for _, img := range repoData.ImgList {
		log.Printf("Pulling image %s (%s) from %s\n", img.Id, img.Tag, remote)
		success := false

		for _, ep := range repoData.Endpoints {
			if err := pullImage(c, img.Id, "https://"+ep+"/v1", repoData.Tokens); err != nil {
				fmt.Printf("Error while retrieving image for tag: %s; checking next endpoint\n", err)
				continue
			}
			success = true
			break
		}

		if !success {
			log.Fatal("Could not find repository on any of the indexed registries.")
		}
	}

	fmt.Fprintf(w, "%v\n", repoData)
}

func setupDocker(r *mux.Router, o Options) {
	context.Path = o.Path
	p := path.Join(context.Path, "containers")
	if err := os.MkdirAll(p, 0700); err != nil && !os.IsExist(err) {
		log.Fatal(err)
		return
	}
	context.Registry = registry.NewRegistry(p)

	p = path.Join(context.Path, "graph")
	if err := os.MkdirAll(p, 0700); err != nil && !os.IsExist(err) {
		log.Fatal(err)
		return
	}
	g, _ := docker.NewGraph(p)
	context.Graph = g

	makeHandler := func(fn func(http.ResponseWriter, *http.Request, *Context)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fn(w, r, &context)
		}
	}

	r.HandleFunc("/registry/pull", makeHandler(pullHandler))
}
