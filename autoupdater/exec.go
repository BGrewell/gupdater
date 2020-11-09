package autoupdater

import (
	"os/exec"
	"strings"
)

// ExecuteCommand executes a single command and returns the results
func ExecuteCommand(command string) (result string, err error) {

	args := strings.Fields(command)
	exe := exec.Command(args[0], args[1:]...)
	out, err := exe.CombinedOutput()
	return string(out), err
}

// ExecuteCommands executes a slice of commands and returns the results
func ExecuteCommands(command []string) (results []string, err error) {

	results = make([]string, len(command))
	for _, c := range command {
		if c == "" {
			continue
		}
		cresult, cerr := ExecuteCommand(c)
		if cerr != nil {
			return results, cerr
		}
		results = append(results, cresult)
	}

	return results, nil
}
