package main

import (
	"fmt"
	"os"
	"strings"
)

// Handler is the function signature of the function to be called when a match
// is found during recursion
type Handler func()

func recurseDirectories(directory string, searchString string, handler Handler) {
	directoryHandle, error := os.Open(directory)
	if error != nil {
		fmt.Printf(errorRecursingDirectories, error)
		os.Exit(1)
	}
	files, error := directoryHandle.Readdir(-1)
	if error != nil {
		fmt.Printf(errorRecursingDirectories, error)
		os.Exit(1)
	}

	searchStringFound := false
	for _, file := range files {

		if file.IsDir() {
			// If this is a directory, recurse into it
			recurseDirectories(fmt.Sprintf("%s/%s", directory, file.Name()), searchString, handler)
		} else {
			// Check each file name to see if we should call the handler
			if searchStringFound == false && strings.Contains(file.Name(), searchString) {
				searchStringFound = true
			}
		}
	}

	if searchStringFound {
		// We found our search string, call the handler
		handler()
	}
}
