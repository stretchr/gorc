package main

import (
	"fmt"
	"os"
	"strings"
)

// callbackHandler is the function signature of the function to be called when a match
// is found during recursion
type callbackHandler func(currentDirectory string)

// skipHandler is the function signature of the function to be called to determine if a directory should be skipped
type skipHandler func(currentDirectory string) bool

func recurseDirectories(directory, targetDirectory string, searchString string, skip skipHandler, callback callbackHandler) {
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
	directoryParts := strings.Split(directory, string(os.PathSeparator))
	directoryName := directoryParts[len(directoryParts)-1]

	for _, file := range files {
		if file.IsDir() && directoryName != targetDirectory {
			if skip(file.Name()) {
				continue
			}
			recurseDirectories(fmt.Sprintf("%s/%s", directory, file.Name()), targetDirectory, searchString, skip, callback)
		} else {
			if searchStringFound == false && strings.Contains(file.Name(), searchString) {
				searchStringFound = true
			}
		}
	}

	if (searchStringFound && targetDirectory == "") || (searchStringFound && directoryName == targetDirectory) {
		// We found our search string, call the handler
		callback(directory)
	}
}
