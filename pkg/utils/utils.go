package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"9fans.net/go/acme"
)

const (
	WIN_ID_SERVICE = "acmefocused"
	VERSION        = "0.0.2-1"
)

func Version() string {
	return VERSION
}
func DailAcmeFocusWin(addr string) (string, error) {
	winid := os.Getenv("winid")
	if winid == "" {
		conn, err := net.Dial("unix", addr)
		if err != nil {
			return "", fmt.Errorf("$winid is empty and could not dial acmefocused: %v", err)
		}
		defer conn.Close()
		b, err := ioutil.ReadAll(conn)
		if err != nil {
			return "", fmt.Errorf("$winid is empty and could not read acmefocused: %v", err)
		}
		return string(bytes.TrimSpace(b)), nil
	}
	return winid, nil
}

func GetCurrentWindow() (*acme.Win, error) {
	winID, err := GetCurrentWinId()
	if err != nil {
		return nil, err
	}

	w, err := acme.Open(winID, nil)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func GetCurrentWinId() (int, error) {
	winStr, err := DailAcmeFocusWin(filepath.Join(getAcmeNamespace(), WIN_ID_SERVICE))
	if err != nil {
		return -1, err
	}

	winID, err := strconv.Atoi(winStr)
	if err != nil {
		return -1, err
	}

	return winID, nil
}

var dotZero = regexp.MustCompile(`\A(.*:\d+)\.0\z`)

// Namespace returns the path to the name space directory for plan9.
func getAcmeNamespace() string {
	ns := os.Getenv("NAMESPACE")
	if ns != "" {
		return ns
	}

	disp := os.Getenv("DISPLAY")
	if disp == "" {
		// No $DISPLAY? Use :0.0 for non-X11 GUI (OS X).
		disp = ":0.0"
	}

	// Canonicalize: xxx:0.0 => xxx:0.
	if m := dotZero.FindStringSubmatch(disp); m != nil {
		disp = m[1]
	}

	// Turn /tmp/launch/:0 into _tmp_launch_:0 (OS X 10.5).
	disp = strings.Replace(disp, "/", "_", -1)

	// NOTE: plan9port creates this directory on demand.
	// Maybe someday we'll need to do that.

	return fmt.Sprintf("/tmp/ns.%s.%s", os.Getenv("USER"), disp)
}
