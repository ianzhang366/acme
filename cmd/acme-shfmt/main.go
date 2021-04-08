package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"9fans.net/go/acme"
)

var version = flag.Bool("v", false, "run gofmt on the entire file after Put")

func main() {
	flag.Parse()

	if *version {
		fmt.Fprintln(os.Stdout, "version: v0.0.8-c")
		return
	}

	l, err := acme.Log()
	if err != nil {
		log.Fatal(err)
	}

	for {
		event, err := l.Read()
		if err != nil {
			log.Fatal(err)
		}
		if event.Name != "" && event.Op == "put" && (strings.HasSuffix(event.Name, ".sh") || strings.HasSuffix(event.Name, "_sh")) {
			reformat(event.ID, event.Name)
		}
	}
}

func reformat(id int, name string) {
	w, err := acme.Open(id, nil)
	if err != nil {
		log.Print(err)
		return
	}
	defer w.CloseFiles()

	//per google style guide https://google.github.io/styleguide/shellguide.html
	o, err := exec.Command("shfmt", "-i", "2", "-sr", "-w", name).CombinedOutput()
	if err != nil {
		w.Errf("shfmt error: %s", err)
		return
	}

	if len(o) != 0 {
		w.Errf("shfmt: %s", string(o))
	}

	w.Ctl("get")
}
