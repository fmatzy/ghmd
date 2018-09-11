package main

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/go-github/github"
)

const (
	usage = `
Usage: ghmd FILE.md
`
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	mdname := os.Args[1]
	ext := filepath.Ext(mdname)
	if ext != ".md" && ext != ".mkd" && ext != ".markdown" {
		fmt.Fprintf(os.Stderr, "%s: invalid file\n", mdname)
		os.Exit(1)
	}
	if _, err := os.Stat(mdname); err != nil {
		fmt.Fprintf(os.Stderr, "%s: not found\n", mdname)
		os.Exit(1)
	}

	tmpl, err := Assets.Open("/assets/index.tmpl")
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(tmpl)
	if err != nil {
		panic(err)
	}
	t := template.Must(template.New("index").Parse(string(b)))

	http.Handle("/assets/style/", http.FileServer(Assets))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		md, err := ioutil.ReadFile(mdname)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		client := github.NewClient(nil)
		body, _, err := client.Markdown(context.Background(), string(md), nil)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 503)
			return
		}
		d := map[string]interface{}{
			"title": mdname,
			"body":  template.HTML(body),
		}
		if err := t.Execute(w, d); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 503)
			return
		}
	})

	addr := ":8080"
	fmt.Fprintf(os.Stderr, "Listening at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
