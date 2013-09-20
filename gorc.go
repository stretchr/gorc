package main

import (
	"fmt"
	"github.com/stretchr/commander"
	"github.com/stretchr/objx"
	"os"
	"os/exec"
)

const (
	CommandInstallTests = iota
	CommandTest
	CommandInstall
	CommandRace
	CommandVet
)

func getwd() string {
	directory, error := os.Getwd()
	if error != nil {
		fmt.Printf(errorCurrentDirectory, error)
		os.Exit(1)
	}
	return directory
}

func execute(command int, packageName string) bool {

	directory := getwd()

	if packageList == nil {
		var fileType string
		if command == CommandTest || command == CommandInstallTests {
			fileType = fileTypeTest
		} else {
			fileType = fileTypeGo
		}
		packageList = buildPackageList(directory, fileType)
		packageList = filterPackages(packageList, packageName, exclusions)
	}

	switch command {
	case CommandInstallTests:
		//fmt.Printf("installing tests...\ngot output: %s\n", runShellCommand(directory, "go", "test", "-i", packageListString))
	case CommandTest:
		runShellCommand(directory, "go", makeArgs(packageList, "test")...)
	case CommandInstall:
	case CommandRace:
	case CommandVet:
	}

	return false

}

// runShellCommand runs a shell command in a specified directory and returns
// a string containing all error output if the command fails
func runShellCommand(directory, command string, arguments ...string) string {
	shellCommand := exec.Command(command, arguments...)
	shellCommand.Dir = directory

	output, _ := shellCommand.CombinedOutput()
	return string(output)

}

func makeArgs(packages []string, commands ...string) []string {

	return append(commands, packages...)

}

var exclusions []string
var packageList []string

func main() {

	var config = readConfig()
	exclusions = config[configKeyExclusions].([]string)

	commander.Go(func() {
		// The default command installs tests, then runs tests.
		commander.Map(commander.DefaultCommand, "", "",
			func(args objx.Map) {
				execute(CommandTest, "")
				return
				if execute(CommandInstallTests, "") {
					execute(CommandTest, "")
				}
			})

		commander.Map("test [packageName=(string)]", "Runs tests, or named test",
			"If no packageName argument is specified, runs all tests recursively. If a packageName argument is specified, runs just that test, unless the argument is \"all\", in which case it runs all tests, including those in the exclusion list.",
			func(args objx.Map) {
				//packageName := args.GetString("packageName")
				fmt.Printf("Found packageNames: %v\n", buildPackageList(getwd(), fileTypeTest))
			})

		commander.Map("install [packageName=(string)]", "Installs tests, or named test",
			"If no packageName argument is specified, installs all tests recursively. If a packageName argument is specified, installs just that test, unless the argument is \"all\", in which case it installs all tests, including those in the exclusion list.",
			func(args objx.Map) {
				//packageName := args.GetString("packageName")
			})

		commander.Map("vet [packageName=(string)]", "Vets packageNames, or named packageName",
			"If no packageName argument is specified, vets all packageNames recursively. If a packageName argument is specified, vets just that packageName, unless the argument is \"all\", in which case it vets all packageNames, including those in the exclusion list.",
			func(args objx.Map) {
				//packageName := args.GetString("packageName")
			})

		commander.Map("race [packageName=(string)]", "Runs race detector on tests, or named test",
			"If no packageName argument is specified, race tests all tests recursively. If a packageName argument is specified, vets just that test, unless the argument is \"all\", in which case it vets all tests, including those in the exclusion list.",
			func(args objx.Map) {
				//packageName := args.GetString("packageName")
			})

		commander.Map("exclude packageName=(string)", "Excludes the named directory from recursion",
			"An excluded directory will be skipped when walking the directory tree. Any subdirectories of the excluded directory will also be skipped.",
			func(args objx.Map) {
				packageName := args.Get("packageName").Str()
				exclude(packageName, config)
				fmt.Printf("\nExcluded \"%s\" from being examined during recursion.\n", packageName)
				config = readConfig()
				exclusions = config[configKeyExclusions].([]string)
				fmt.Printf("\n%s\n\n", formatExclusionsForPrint(exclusions))
			})

		commander.Map("include packageName=(string)", "Removes the named directory from the exclusion list", "",
			func(args objx.Map) {
				packageName := args.Get("packageName").Str()
				include(packageName, config)
				fmt.Printf("\nRemoved \"%s\" from the exclusion list.\n", packageName)
				fmt.Printf("\n%s\n\n", formatExclusionsForPrint(exclusions))
			})

		commander.Map("exclusions", "Prints the exclusion list", "",
			func(args objx.Map) {
				fmt.Printf("\n%s\n\n", formatExclusionsForPrint(exclusions))
			})
	})

}
