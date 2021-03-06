package main

// code is from https://github.com/sminez/acme-corp/
// put all together to minimize the dependency and use https://github.com/fhs/acme-lsp/blob/master/cmd/acmefocused to get current winid, since I'm already using acmefocused with acme-lsp
/*
pick - a minimalist input selector for the acme text editor modelled after dmenu

If launched from within acme itself, pick will use the current acme window as defined by the 'winid'
environment variable. Otherwise, it will attempt to query a running snooper instance to fetch the
focused window id.
  * To mimic dmenu behaviour of reading input from stdin, pass the '-s' flag.
  * To return the index of the selected line in the input instead of the line itself, pass the '-n' flag.
  * To override the default prompt ('> ') pass the '-p' flag followed by the string to use as the prompt.

+pick window actions
  * character input will be interpreted as a regex for filtering lines
  * button 3: select a line to return
  * Return:   select the line the curesor is currently on
*/

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"9fans.net/go/acme"
	"github.com/ianzhang366/acme/pkg/utils"
)

const (
	lineOffset = 1 // assuming single line prompt
	windowName = "+pick"
	VERSION    = "0.0.2"
)

var (
	readFromStdIn = flag.Bool("s", false, "read input from stdin instead of the current acme window")
	returnLineNum = flag.Bool("n", false, "return the line number of the selected line, not the line itself")
	numberLines   = flag.Bool("N", false, "prefix each line with its line number")
	prompt        = flag.String("p", "> ", "prompt to present to the user when taking input")
	version       = flag.Bool("v", false, "prefix each line with its line number")
)

type linePicker struct {
	w              *acme.Win
	rawLines       []string
	lineMap        map[int]string // original line numbers
	selectedLines  map[int]int    // window line numbers -> input line number
	currentInput   string
	selectionEvent *acme.Event
}

func newLinePicker(rawLines []string) *linePicker {
	var w *acme.Win
	var err error

	if w, err = acme.New(); err != nil {
		fmt.Printf("Unable to initialise new acme window: %s\n", err)
		os.Exit(1)
	}

	w.Name(windowName)
	lineMap := make(map[int]string)

	for n, l := range rawLines {
		lineMap[n+1] = l
	}

	return &linePicker{
		w:             w,
		rawLines:      rawLines,
		lineMap:       lineMap,
		selectedLines: make(map[int]int),
		currentInput:  "",
	}
}

func (lp *linePicker) selectedLine() (int, string, error) {
	var (
		windowLineNumber int
		err              error
	)

	if windowLineNumber, err = EventLineNumber(lp.w, lp.selectionEvent); err != nil {
		return -1, "", err
	}

	// hitting enter on the input line selects top match if there is at least one, otherwise
	// it returns the current input text
	if windowLineNumber == 0 {
		if len(lp.selectedLines) == 0 {
			return -1, lp.currentInput, nil
		}
		windowLineNumber = 1
	}

	lineNumber := lp.selectedLines[windowLineNumber-lineOffset]
	return lineNumber, lp.lineMap[lineNumber], nil
}

func (lp *linePicker) filter() (int, string, error) {
	ef := &EventFilter{
		KeyboardInputBody: func(w *acme.Win, e *acme.Event, done func() error) error {
			r := e.Text[0]

			if r <= 26 {
				switch fmt.Sprintf("C-%c", r+96) {
				case "C-j": // (Enter)
					lp.selectionEvent = e
					return done()
				case "C-d":
					lp.w.Del(true)
					os.Exit(0)
				}
			}

			lp.currentInput += string(r)
			return lp.reRender()
		},

		KeyboardDeleteBody: func(w *acme.Win, e *acme.Event, done func() error) error {
			if l := len(lp.currentInput); l > 0 {
				removed := e.Q1 - e.Q0
				lp.currentInput = lp.currentInput[:l-removed]
			}

			return lp.reRender()
		},

		Mouse3Body: func(w *acme.Win, e *acme.Event, done func() error) error {
			lp.selectionEvent = e
			return done()
		},
	}

	if err := lp.reRender(); err != nil {
		return -1, "", err
	}

	if err := ef.Filter(lp.w); err != nil {
		return -1, "", err
	}

	return lp.selectedLine()
}

func (lp *linePicker) reRender() error {
	lines := lp.rawLines
	lp.w.Clear()

	if len(lp.currentInput) > 0 {
		fragments := strings.Split(lp.currentInput, " ")
		lp.selectedLines = make(map[int]int)
		lines = []string{}
		k := 0

		for ix, line := range lp.lineMap {
			if containsAll(line, fragments) {
				lp.selectedLines[k] = ix
				lines = append(lines, line)
				k++
			}
		}
	}

	lp.w.Write("body", []byte(fmt.Sprintf("%s%s\n", *prompt, lp.currentInput)))
	lp.w.Write("body", []byte(strings.Join(lines, "\n")))
	SetCursorEOL(lp.w, 1)
	return nil
}

func containsAll(s string, ts []string) bool {
	for _, t := range ts {
		if !strings.Contains(s, t) {
			return false
		}
	}
	return true
}

func readFromAcme() ([]string, error) {
	var (
		w   *acme.Win
		err error
	)

	if w, err = utils.GetCurrentWindow(); err != nil {
		return nil, err
	}
	defer w.CloseFiles()

	return WindowBodyLines(w)
}

func numberedLines(lines []string) []string {
	for ix, line := range lines {
		lines[ix] = fmt.Sprintf("%3d | %s", ix+1, line)
	}

	return lines
}

func main() {
	var (
		lines     []string
		err       error
		n         int
		selection string
	)

	flag.Parse()

	if *version {
		fmt.Fprintf(os.Stdout, "pick service version: %s \n", utils.Version())
		os.Exit(0)
	}

	if *readFromStdIn {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			lines = append(lines, s.Text())
		}
	} else {
		if lines, err = readFromAcme(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if *numberLines {
		lines = numberedLines(lines)
	}

	lp := newLinePicker(lines)
	defer lp.w.Del(true)

	if n, selection, err = lp.filter(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *returnLineNum {
		fmt.Println(n)
	} else {
		fmt.Println(selection)
	}
}

// A handler fucntion that processes an Acme event and takes an action. Passthrough must be explicitly
// carried out by the handler function itself.
type handler = func(*acme.Win, *acme.Event, func() error) error

// An EventFilter takes hold of an acme window's event file and passes all events
// it sees through a set of filter functions if they are defined. Unmatched events
// are passed through to acme.
type EventFilter struct {
	complete           bool
	KeyboardInputBody  handler
	KeyboardDeleteBody handler
	KeyboardInputTag   handler
	KeyboardDeleteTag  handler
	Mouse2Body         handler
	Mouse3Body         handler
	Mouse2Tag          handler
	Mouse3Tag          handler
}

func (ef *EventFilter) markComplete() error {
	ef.complete = true
	return nil
}

func (ef *EventFilter) applyOrPassthrough(f handler, w *acme.Win, e *acme.Event) error {
	if f != nil {
		return f(w, e, ef.markComplete)
	}

	return w.WriteEvent(e)
}

// Filter runs the event filter, releasing the window event file on the first error encountered
func (ef *EventFilter) Filter(w *acme.Win) error {
	for e := range w.EventChan() {
		if err := ef.filterSingle(w, e); err != nil {
			return err
		}

		if ef.complete {
			return nil
		}
	}

	return fmt.Errorf("lost event channel")
}

// Currently dropping E and F events that are generated by writes from other programs to the acme
// control files.
func (ef *EventFilter) filterSingle(w *acme.Win, e *acme.Event) error {
	switch e.C1 {
	case 'K':
		switch e.C2 {
		case 'I':
			ef.applyOrPassthrough(ef.KeyboardInputBody, w, e)
		case 'D':
			ef.applyOrPassthrough(ef.KeyboardDeleteBody, w, e)
		case 'i':
			ef.applyOrPassthrough(ef.KeyboardInputTag, w, e)
		case 'd':
			ef.applyOrPassthrough(ef.KeyboardDeleteTag, w, e)
		}
		return nil

	case 'M':
		switch e.C2 {
		case 'X':
			return ef.applyOrPassthrough(ef.Mouse2Body, w, e)
		case 'L':
			return ef.applyOrPassthrough(ef.Mouse3Body, w, e)
		case 'x':
			return ef.applyOrPassthrough(ef.Mouse2Tag, w, e)
		case 'l':
			return ef.applyOrPassthrough(ef.Mouse3Tag, w, e)
		}
	}

	w.WriteEvent(e)
	return nil
}

// SetCursorEOL will position the current window cursor at the end of line.
func SetCursorEOL(w *acme.Win, line int) {
	w.Addr(fmt.Sprintf("%d-#1", line+1))
	w.Ctl("dot=addr")
	w.Ctl("show")
}

// SetCursorBOL will position the current window cursor at the beginning of line.
func SetCursorBOL(w *acme.Win, line int) {
	w.Addr(fmt.Sprintf("%d-#0", line))
	w.Ctl("dot=addr")
	w.Ctl("show")
}

// WindowBody reads the body of the current window as single string
func WindowBody(w *acme.Win) (string, error) {
	var (
		body []byte
		err  error
	)

	// TODO: stash and restore current addr
	w.Addr(",")
	if body, err = w.ReadAll("data"); err != nil {
		return "", err
	}
	return string(body), nil
}

// WindowBodyLines reads the body of the current window as an array of strings split on newline
func WindowBodyLines(w *acme.Win) ([]string, error) {
	body, err := WindowBody(w)
	if err != nil {
		return nil, err
	}
	return strings.Split(body, "\n"), nil
}

// EventLineNumber returns the line that a e occurred on in w
func EventLineNumber(w *acme.Win, e *acme.Event) (int, error) {
	body, err := WindowBody(w)
	if err != nil {
		return -1, err
	}

	upToCursor := body[:e.Q0]
	return strings.Count(upToCursor, "\n"), nil
}
