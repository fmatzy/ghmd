package main

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
)

const (
	usage = `
Usage: ghmd FILE.md
`

	index = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>{{ .title }}</title>
<link rel="stylesheet" href="/assets/github-markdown.css" media="all">
</head>
<body>
<div class="markdown-body">{{ .body }}</div>
</body>
</html>
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
	_, err := os.Stat(mdname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: not found\n", mdname)
		os.Exit(1)
	}

	t := template.Must(template.New("index").Parse(index))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path
		if strings.HasPrefix(name, "/assets/") {
			b, err := ioutil.ReadFile(name[1:])
			if err != nil {
				http.NotFound(w, r)
				return
			}

			w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(name)))
			w.Write(b)
			return
		}

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
