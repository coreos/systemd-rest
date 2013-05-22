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
	"regexp"
)

const ContainerDir = "/var/lib/containers/"

type Context struct {
	Path string
	Registry *registry.Registry
	Graph *docker.Graph
	Repositories *docker.TagStore
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

// TODO: add tag support
func pullHandler(w http.ResponseWriter, r *http.Request, c *Context) {
	vars := mux.Vars(r)
	remote := vars["remote"]

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

	for tag, id := range tagsList {
		if err := c.Repositories.Set(remote, tag, id, true); err != nil {
			log.Fatal(err)
			return
		}
	}
	if err := c.Repositories.Save(); err != nil {
		log.Fatal(err)
		return
	}

	fmt.Fprintf(w, "%v\n", repoData)
}

func createHandler(w http.ResponseWriter, r *http.Request, c *Context) {
	imageName := r.FormValue("image")

	// TODO: @philips Don't hardcode the tag name here
	image, err := c.Repositories.GetImage(imageName, "latest")
	if err != nil {
		log.Fatal(err)
		return
	}

	if imageName == "" {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Cannot find container image: %s", imageName)
		return
	}

	// Figure out the resting place of the container
	vars := mux.Vars(r)
	container := vars["container"]

	validID := regexp.MustCompile(`^[A-Za-z0-9]+$`)
	if !validID.MatchString(container) {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Invalid container name: %s\n", container)
		return
	}

	container = path.Join(ContainerDir, container)

	err = os.Mkdir(container, 0700)
	if os.IsExist(err) {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Existing container: %s\n", container)
		return
	}
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, "Error")
		log.Fatal(err)
		return
	}

	copyAll := func(img *docker.Image) (err error) {
		tarball, err := image.TarLayer(docker.Uncompressed)
		if err != nil {
			return err
		}
		if err := docker.Untar(tarball, container); err != nil {
			return err
		}
		return
	}

	err = image.WalkHistory(copyAll)

	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Fprint(w, "ok")
	return
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

	p = path.Join(context.Path, "repositories")
	t, _ := docker.NewTagStore(p, g)
	context.Repositories = t

	makeHandler := func(fn func(http.ResponseWriter, *http.Request, *Context)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fn(w, r, &context)
		}
	}

	r.HandleFunc("/registry/pull/{remote:.*}", makeHandler(pullHandler))
	r.HandleFunc("/container/create/{container:.*}", makeHandler(createHandler))
}
