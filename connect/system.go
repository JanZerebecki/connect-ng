package connect

import (
	"bytes"
	"os"
	"os/exec"
)

func execute(cmd []string, quiet bool, validExitCodes []int) ([]byte, error) {
	Debug.Printf("Executing: %s Quiet: %v\n", cmd, quiet)
	var stderr, stdout bytes.Buffer
	comm := exec.Command(cmd[0], cmd[1:]...)
	comm.Stdout = &stdout
	comm.Stderr = &stderr
	comm.Env = append(os.Environ(), "LC_ALL=C")
	err := comm.Run()
	exitCode := comm.ProcessState.ExitCode()
	Debug.Printf("Return code: %d\n", exitCode)
	if stdout.Len() > 0 {
		Debug.Println("Output:", stdout.String())
	}
	if stderr.Len() > 0 {
		Debug.Println("Error:", stderr.String())
	}
	// TODO Ruby version also checks stderr for "ABORT request"
	if err != nil && !containsInt(validExitCodes, exitCode) {
		output := stderr.Bytes()
		// zypper with formatter option writes to stdout instead of stderr
		if len(output) == 0 {
			output = stdout.Bytes()
		}
		ee := ExecuteError{Commmand: cmd, ExitCode: exitCode, Output: output, Err: err}
		return nil, ee
	}
	if quiet {
		return nil, nil
	}
	return stdout.Bytes(), nil
}

func containsInt(s []int, i int) bool {
	for _, e := range s {
		if e == i {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
