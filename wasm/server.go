// +build disabled

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func main() {
	const wasm = "/doodle.wasm"

	// Watch the dev directory for changes.
	go watchChanges()

	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc(wasm, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/wasm")
		http.ServeFile(w, r, "."+wasm)
	})

	fmt.Println("Listening at http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// onChange handler to rebuild the wasm file automatically.
func onChange() {
	out, err := exec.Command("make").Output()
	if err != nil {
		log.Printf("error: %s", err)
	}
	log.Printf("%s", out)
	log.Printf("Doodle WASM file rebuilt")
}

// Watch the Doodle source tree for changes to Go files and rebuild
// the wasm binary automatically.
func watchChanges() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("error: %s", err)
		return
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		log.Println("Starting watch files loop")
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				log.Printf("event: %s", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("modified file: %s", event.Name)
					onChange()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				log.Printf("error: %s", err)
			}
		}
	}()

	log.Println("Adding source directory to watcher")
	dirs := crawlDirectory("../")

	// Watch all these folders.
	for _, dir := range dirs {
		err = watcher.Add(dir)
		if err != nil {
			log.Printf("error: %s", err)
		}
	}
	<-done
}

// Crawl the filesystem and return paths with Go files.
func crawlDirectory(root string) []string {
	var (
		ext    = ".go"
		result []string
		has    bool
	)

	files, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatalln(err)
	}

	for _, file := range files {
		if file.Name() == ".git" {
			continue
		}

		// Recursively scan subdirectories.
		if file.IsDir() {
			result = append(result,
				crawlDirectory(filepath.Join(root, file.Name()))...,
			)
			continue
		}

		// This root has a file we want?
		if filepath.Ext(file.Name()) == ext {
			has = true
		}
	}

	if has {
		result = append(result, root)
	}

	return result
}
