package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// sliceContainsString determines if a slice of string contains the target string
func sliceContainsString(target string, slice []string) (bool, int) {
	for index, value := range slice {
		if value == target {
			return true, index
		}
	}
	return false, -1
}

// stringSliceFromInterfaceSlice creates a []string from a []interface{}
func stringSliceFromInterfaceSlice(interfaceSlice []interface{}) []string {
	retval := make([]string, len(interfaceSlice))
	for i, str := range interfaceSlice {
		retval[i] = str.(string)
	}
	return retval
}

// formatExclusionsForPrint returns a string detailing all excluded directories.
func formatExclusionsForPrint(exclusions []string) string {

	excludedPackages := strings.Join(exclusions, "\n\t")
	return fmt.Sprintf("Excluded Directories:\n\t%s", excludedPackages)

}

// runShellCommand runs a shell command in a specified directory and returns
// a string containing all error output if the command fails
func runShellCommand(directory, command string, arguments ...string) string {
	shellCommand := exec.Command(command, arguments...)
	shellCommand.Dir = directory

	if output, error := shellCommand.CombinedOutput(); error != nil {
		return string(output)
	}

	return ""
}
