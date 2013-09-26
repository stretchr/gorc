package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"strings"
	"path"
)

// filterPackages filters packages based on the named package and the exclusion list
func filterPackages(packages []string, target string, exclusions []string) []string {

	if target == "all" {
		return packages
	}

	filteredPackages := make([]string, 0, len(packages))

	for _, pkg := range packages {
		if target != "" {
			if strings.Contains(pkg, target) {
				filteredPackages = append(filteredPackages, pkg)
				break
			}
		} else {
			if contains, _ := stringInSlice(pkg, exclusions); !contains {
				filteredPackages = append(filteredPackages, pkg)
			}
		}
	}

	return filteredPackages

}

// buildPackageList recurses through a directory tree and returns a list of packages for the go command to act upon
// It also filters the results to only include packages that contain the type of file we are looking for
func buildPackageList(startingDirectory, fileType string) []string {
	packages := make([]string, 0)
	_buildPackageList(startingDirectory, fileType, &packages)

	for i, pkg := range packages {
		packages[i] = strings.Replace(pkg, startingDirectory, ".", 1)
	}

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

// buildSubdirList recurses through a directory tree and returns a
// map of sub-directories.  This doesn't do any checking for packages
// or other go files - it's mainly for the watch command, which needs
// to watch all sub-directories, regardless of contents.  A map is
// used to make removal of directories easier.
func buildSubdirMap(directory string) map[string]bool {
	subdirs := make(map[string]bool)
	if _, err := os.Stat(directory); err == nil {
		_buildSubdirMap(directory, subdirs)
	}
	return subdirs
}

// _buildSubdirList is the recursive helper for buildSubdirMap]
func _buildSubdirMap(directory string, subdirs map[string]bool) {
	fmt.Printf("Adding directory %s\n", directory)
	subdirs[directory] = true
	dirs, err := ioutil.ReadDir(directory)
	if err != nil {
		panic(err)
	}
	for _, dir := range dirs {
		if dir.IsDir() && !strings.HasPrefix(dir.Name(), ".") {
			fmt.Printf("Found directory %s\n", dir.Name())
			fullPath := path.Join(directory, dir.Name())
			fmt.Printf("Full path %s\n", fullPath)
			_buildSubdirMap(fullPath, subdirs)
		}
	}
}
