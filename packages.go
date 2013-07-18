package main

import (
	"fmt"
	"os"
	"strings"
)

// filterPackages filters packages based on the named package and the exclusion list
func filterPackages(packages []string, target string, exclusions []string) []string {

	if target == "all" {
		return packages
	}

	filteredPackages := make([]string, 0, len(packages))

	if target != "" {
		for _, pkg := range packages {
			if strings.Contains(pkg, target) {
				filteredPackages = append(filteredPackages, pkg)
				break
			}
		}
	}

}

// buildPackageList recurses through a directory tree and returns a list of packages for the go command to act upon
// It also filters the results to only include packages that contain the type of file we are looking for
func buildPackageList(startingDirectory, fileType string) []string {
	packages := make([]string, 0)
	_buildPackageList(startingDirectory, fileType, &packages)
	return packages
}

// _buildPackageList recurses through a directory tree and returns a list of packages for the go command to act upon
// It also filters the results to only include packages that contain the type of file we are looking for
func _buildPackageList(directory, fileType string, packages *[]string) {
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

	fileTypeFound := false

	for _, file := range files {
		if file.IsDir() {
			if strings.HasPrefix(file.Name(), ".") {
				continue
			}
			_buildPackageList(fmt.Sprintf("%s/%s", directory, file.Name()), fileType, packages)
		} else {
			if fileTypeFound == true {
				continue
			}
			if strings.Contains(file.Name(), fileType) {
				fileTypeFound = true
			}
		}
	}

	if fileTypeFound {
		*packages = append(*packages, directory)
	}

}
