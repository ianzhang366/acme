package main

import (
	"log"
	"os/exec"
	"strings"

	"9fans.net/go/acme"
)

func main() {
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

	_, err = exec.Command("shfmt", "-sr", "-w", name).CombinedOutput()
	if err != nil {
		w.Errf("shfmt error: %s", err)
		return
	}

	w.Ctl("get")
}
