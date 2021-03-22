package main

import (
	"flag"
	"io/ioutil"
	"log"
	"strings"

	"9fans.net/go/acme"
)

var defaultTagFile string

var tagFile = flag.String("f", "./tags", "append tags from the given file to acme windows")

func main() {
	flag.Parse()
	l, err := acme.Log()
	if err != nil {
		log.Fatal(err)
	}

	tags, err := ioutil.ReadFile(*tagFile)
	if err != nil {
		log.Fatalf("failed to log tag files, err: %v", err)
	}

	for {
		event, err := l.Read()
		if err != nil {
			log.Fatal(err)
		}

		//event.Op: focus, get, del, put
		if event.Name != "" && event.Op == "focus" {
			appendTags(event.ID, event.Name, tags)
		}
	}
}

func appendTags(id int, name string, tags []byte) {
	w, err := acme.Open(id, nil)
	if err != nil {
		log.Print("failed to open file", err)
		return
	}

	defer w.CloseFiles()

	curTag, err := w.ReadAll("tag")
	if err != nil {
		log.Print(err)
		return
	}

	if strings.Contains(string(curTag), string(tags)) {
		return
	}

	w.Write("tag", tags)
}
