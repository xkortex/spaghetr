package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// ScanLines is a split function for a Scanner that returns each line of
// text, stripped of any trailing end-of-line marker. The returned line may
// be empty. The end-of-line marker is one optional carriage return followed
// by one mandatory newline. In regular expression notation, it is `(\r\n|\r|\n`.
// The last non-empty line of input will be returned even if it has no
// newline.
func ScanRLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	// todo: swap with regex
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i+1], nil
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// We have a \r terminated line
		return i + 1, data[0:i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func main() {
	cmd := exec.Command("python", "./spaghetr/aux/status_bar_dummy.py")

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	scanner_out := bufio.NewScanner(stdout)
	scanner_err := bufio.NewScanner(stderr)
	scanner_out.Split(ScanRLines)
	scanner_err.Split(ScanRLines)
	cmd.Start()
	go func() {
		bout := bufio.NewWriter(os.Stdout)
		berr := bufio.NewWriter(os.Stderr)
		for {
			//statOut := scanner_out.Scan()
			//statErr := scanner_err.Scan()
			if !scanner_out.Scan() && !scanner_err.Scan() {
				break
			}
			//if !statOut && !statErr {
			//	break
			//}
			o_msg := scanner_out.Bytes()
			e_msg := scanner_err.Bytes()
			//bout.Write([]byte("."))
			bout.Write(o_msg)
			berr.Write(e_msg)
			bout.Flush()
			berr.Flush()
		}
	}()
	fmt.Println("waiting")
	cmd.Wait()
	fmt.Println("done")

}
