package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"9fans.net/go/acme"
)

var version = flag.Bool("v", false, "run gofmt on the entire file after Put")

// want to call this within the following case
// 1. running in background watch for acme script files and shfmt on put
// we can read all files and parse the first line
// 2. call by script and mirror the default shfmt behaviour
// have shell flag then pass the rest

func main() {
	flag.Parse()

	if *version {
		fmt.Fprintln(os.Stdout, "version: v0.0.8-g")
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
		if event.Name != "" && event.Op == "put" && (strings.HasSuffix(event.Name, ".sh") || (filepath.Ext(event.Name) == "")) {
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

	firstLine := make([]byte, 512)
	// here the read is read from one of acme's window file, tag, data, boday...
	_, err = w.Read("body", firstLine)
	if err != nil {
		w.Errf("fail to read the first line of file, error %v", err)
		return
	}

	// pre google shell guide
	// if the script is lib, then the file name ends with .sh
	// if the script is executable, then the file will start with sengbang without ext .sh
	if !strings.HasSuffix(name, ".sh") && !strings.HasPrefix(string(firstLine), "#!/bin/bash") {
		return
	}

	//per google style guide https://google.github.io/styleguide/shellguide.html
	// err is the command return code.
	o, err := exec.Command("shfmt", "-i", "4", "-sr", "-w", name).CombinedOutput()
	if err != nil {
		w.Errf("shfmt: %serror: %v\n", string(o), err)
		return
	}

	w.Ctl("get")
}
