package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gomarkdown/markdown"
)

const (
	usage = `
Usage: ghmd [-t TEMPLATE] FILE.md
`
)

var (
	tOpt = flag.String("t", "", "template file")
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	mdname := flag.Arg(0)
	ext := filepath.Ext(mdname)
	if ext != ".md" && ext != ".mkd" && ext != ".markdown" {
		fmt.Fprintf(os.Stderr, "%s: invalid file\n", mdname)
		os.Exit(1)
	}
	if _, err := os.Stat(mdname); err != nil {
		fmt.Fprintf(os.Stderr, "%s: not found\n", mdname)
		os.Exit(1)
	}

	var tmpl io.Reader
	var err error
	if *tOpt != "" {
		tmpl, err = os.Open(*tOpt)
	} else {
		tmpl, err = Assets.Open("/assets/index.tmpl")
	}
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

		body := markdown.ToHTML(md, nil, nil)
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
