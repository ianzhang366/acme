// Package acme extends the 9fans.net/go/acme
// package with additional functionality.

// source code comes from:
// https://github.com/s-urbaniak/acme/blob/master/acme.go
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"9fans.net/go/acme"
	"github.com/ianzhang366/acme/pkg/utils"
)

var (
	version = flag.Bool("v", false, "prefix each line with its line number")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Fprintf(os.Stdout, "line number service version: %s \n", utils.Version())
		os.Exit(0)
	}

	winID, err := utils.GetCurrentWinId()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return
	}

	win, err := utils.GetCurrentWindow()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return
	}

	ln, err := LineAddress(winID, win)

	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return
	}

	fileN, err := getFilename(win)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return
	}

	fmt.Fprintf(os.Stdout, "%s:%v\n", fileN, ln)
}

// LineAddress returns the line address
// at the current cursor position.
func LineAddress(winId int, win *acme.Win) (int, error) {
	_, _, err := win.ReadAddr() // make sure address file is already open.
	if err != nil {
		return -1, fmt.Errorf("cannot read address: %v", err)
	}
	err = win.Ctl("addr=dot")
	if err != nil {
		return -1, fmt.Errorf("cannot set addr=dot: %v", err)
	}
	q0, _, err := win.ReadAddr()
	if err != nil {
		return -1, fmt.Errorf("cannot read address: %v", err)
	}

	body, err := readBody(winId)
	if err != nil {
		return -1, fmt.Errorf("cannot read body: %v", err)
	}
	return 1 + nlcount(body, q0), nil
}

func nlcount(b []byte, q0 int) int {
	nl := 0
	ri := 0
	for _, r := range string(b) {
		if ri == q0 {
			return nl
		}
		if r == '\n' {
			nl++
		}
		ri++
	}
	return nl
}

// ReadBody returns the text body content.
func readBody(winid int) ([]byte, error) {
	rwin, err := acme.Open(winid, nil)

	if err != nil {
		return nil, err
	}

	defer rwin.CloseFiles()

	var body []byte
	buf := make([]byte, 8000)
	for {
		n, err := rwin.Read("body", buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		body = append(body, buf[0:n]...)
	}

	if err != nil {
		return nil, err
	}

	return body, nil
}

// Filename returns the file name.
func getFilename(win *acme.Win) (string, error) {
	tagb, err := win.ReadAll("tag")
	if err != nil {
		return "", fmt.Errorf("cannot read tag: %v", err)
	}

	tag := string(tagb)
	i := strings.Index(tag, " ")
	if i == -1 {
		return "", fmt.Errorf("tag contains no spaces")
	}

	return tag[0:i], nil
}
