package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"9fans.net/go/acme"
)

const (
	Sep = "\n"
)

var tagFile = flag.String("f", "./tags", "append tags from the given file to acme windows")

func main() {
	flag.Parse()
	l, err := acme.Log()
	if err != nil {
		log.Fatal(err)
	}

	ts, err := ioutil.ReadFile(*tagFile)
	if err != nil {
		log.Fatalf("failed to log tag files, err: %v", err)
	}

	tags := strings.Split(string(ts), Sep)
	if len(tags) == 0 {
		return
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

func appendTags(id int, name string, tags []string) {
	w, err := acme.Open(id, nil)
	if err != nil {
		log.Print("failed to open file ", err)
		return
	}

	defer w.CloseFiles()

	if w == nil {
		log.Print("it seems we got an empty winid")
		return
	}

	curTag := make([]byte, 1024)
	_, err = w.Read("tag", curTag)
	//	curTag, err := w.ReadAll("tag")
	if err != nil {
		log.Print(err)
		return
	}

	fs := strings.Split(string(curTag), " ")

	if strings.HasSuffix(fs[0], "+watch") {
		return
	}

	for _, tag := range tags {
		if strings.Contains(string(curTag), string(tag)) {
			continue
		}

		nTag := fmt.Sprintf(" %s", tag)
		w.Write("tag", []byte(nTag))
	}
}
