package main

import (
	"fmt"
	"strings"
)

// stringInSlice determines if a slice of string contains the target string
func stringInSlice(target string, slice []string) (bool, int) {
	for index, value := range slice {
		if strings.Contains(target, value) {
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
