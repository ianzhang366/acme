// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Editinacme can be used as $EDITOR in a Unix environment.
//
// Usage:
//
//	editinacme <file>
//
// Editinacme uses the plumber to ask acme to open the file,
// waits until the file's acme window is deleted, and exits.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

//I want the make this one
//1. able to read from stdin, then open a window on acme for it
//2. able to handle file name as input

func main() {
	log.SetFlags(0)
	log.SetPrefix("editinacme: ")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: editinacme file\n")
		os.Exit(2)
	}
	flag.Parse()
	if flag.NArg() > 1 {
		flag.Usage()
	}

	file := flag.Arg(0)

	//standard input flow
	if file == "" || file == "-" {
		fmt.Println("hi standard input flag is in use")
		file = "/dev/stdin"
		cmd := exec.Command("plumb", "-i", "-d", "edit", "-a", "filename=+Errors")
		cmd.Stdin = os.Stdin
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("executing plumb: %v\n%s", err, out)
		}

		return
	}

	_, err := os.Stat(file)
	if err != nil {
		log.Fatal(err)
	}

	fullpath, err := filepath.Abs(file)
	if err != nil {
		log.Fatal(err)
	}
	file = fullpath

	//	r, err := acme.Log()
	//	if err != nil {
	//		log.Fatal(err)
	//	}

	log.Printf("editing %s", file)

	out, err := exec.Command("plumb", "-d", "edit", file).CombinedOutput()
	if err != nil {
		log.Fatalf("executing plumb: %v\n%s", err, out)
	}

	//	for {
	//		ev, err := r.Read()
	//		if err != nil {
	//			log.Fatalf("reading acme log: %v", err)
	//		}
	//		if ev.Op == "del" && ev.Name == file {
	//			break
	//		}
	//	}

}