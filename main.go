package run

import (
	"bufio"
	"fmt"
	"github.com/xkortex/spaghetr/spaghetr"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("python", "./spaghetr/aux/status_bar_dummy.py")

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	scanner_out := bufio.NewScanner(stdout)
	scanner_err := bufio.NewScanner(stderr)
	scanner_out.Split(spaghetr.ScanRLines)
	scanner_err.Split(spaghetr.ScanRLines)
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
