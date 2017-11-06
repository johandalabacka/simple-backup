package main

import (
	"os/exec"
	"syscall"
)

func RunCommand(command string, arguments ...string) (string, int) {
	cmd := exec.Command(command, arguments...)
	raw, err := cmd.CombinedOutput()
	output := string(raw)

	var exitCode int
	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			// This will happen (in OSX) if `name` is not available in $PATH,
			// in this situation, exit code could not be get, and stderr will be
			// empty string very likely, so we use the default fail code, and format err
			// to string and set to stderr
			// log.Printf("Could not get exit code for failed program: %v, %v", name, args)
			exitCode = 1
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	return output, exitCode
}
