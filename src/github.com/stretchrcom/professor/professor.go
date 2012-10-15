package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// kArgumentErrorUsage is printed when no arguments are passed to the program.
var kArgumentErrorUsage string = `Professor run or installs test dependencies, recursively, starting from the current working directory

Usage: professor command [subcommand]

Valid commands are:

	run - Runs tests
	run [all] - Runs tests, including excluded directories
	install - Installs dependencies for tests
	exclude <directory name> - Excludes a directory from recursion
	include <directory name> - Includes a directory in recursion
	exclusions - Prints a list of excluded directories

Version 1.0`

var (
	// kArgumentErrorUnknownCommand is printed when an unknown command is passed to the program.
	kArgumentErrorUnknownCommand string = "Unknown command: %s"

	// kArgumentErrorUnknownSubcommandCommand is printed when an unknown subcommand is passed to the program.
	kArgumentErrorUnknownSubcommand string = "Unknown subcommand: %s"

	// kArgumentErrorSubcommandRequired is printed when a subcommand is required, but was not passed.
	kArgumentErrorSubcommandRequired string = "%s requires a subcommand"

	// kErrorSavingFile is printed when an error occurs attempting to save the configuration file.
	kErrorSavingFile = "There was an error attempting to save your configuration file."

	// kErrorRecursingDirectories is printed when an error occurs recursing through the directory structure.
	kErrorRecursingDirectories = "There was an error when attempting to recurse directories: %s"

	// kErrorCurrentDirectory is printed when an error occurs attempting to get the current working directory.
	kErrorCurrentDirectory = "There was an error attempting to get directory in which professor is being run: %s"

	// kCommandRun is the string for the run command.
	kCommandRun string = "run"

	// kCommandInstall is the string for the install command.
	kCommandInstall string = "install"

	// kCommandExclude is the string for the exclude command.
	kCommandExclude string = "exclude"

	// kCommandInclude is the string for the include command.
	kCommandInclude string = "include"

	// kCommandExclusions is the string for the exclusions command.
	kCommandExclusions string = "exclusions"

	// kValidCommands contains the valid top level commands. Used to verify the top level command is sane.
	kValidCommands = []string{kCommandRun, kCommandInstall, kCommandExclude, kCommandInclude, kCommandExclusions}

	// kCommandsRequiringSubcommands contains the top level commands that require a subcommand. Used to enforce subcommands when required.
	kCommandsRequiringSubcommands = []string{kCommandExclude, kCommandInclude}

	// kSubcommandAll is the string for the all subcommand. Used with the run command.
	kSubcommandAll string = "all"

	// kValidSubcommands contains the valid subcommands. Used to verify subcommand is sane.
	kValidSubcommands = []string{kSubcommandAll}

	// kConfigKeyExclusions is the string for the key in the configuration object at which the exclusions list is stored
	kConfigKeyExclusions = "exclusions"

	// kConfigFilename is the string for the name of the professor configuration file
	kConfigFilename = ".professor"

	// kShellCommandInstallDependencies is the string for the shell command to run when installing dependencies
	kShellCommandInstallDependencies string = "go test -i"

	// kShellCommandRunTest is the string for the shell command to run when executing tests
	kShellCommandRunTest string = "go test"
)

// SliceContainsString determines if a slice of string contains the target string
func SliceContainsString(target string, slice []string) (bool, int) {
	for index, value := range slice {
		if value == target {
			return true, index
		}
	}
	return false, -1
}

// StringSliceFromInterfaceSlice creates a []string from a []interface{}
func StringSliceFromInterfaceSlice(interfaceSlice []interface{}) []string {
	retval := make([]string, len(interfaceSlice))
	for i, str := range interfaceSlice {
		retval[i] = str.(string)
	}
	return retval
}

// VerifyArguments verifies that the arguments passed are sane
// If the arguments are not sane, returns false and a string detailing the proper usage.
func VerifyArguments(arguments []string) (bool, string) {

	if len(arguments) == 1 || len(arguments) > 3 {
		return false, kArgumentErrorUsage
	}

	// Verify that the command is valid
	command := arguments[1]
	if contains, _ := SliceContainsString(command, kCommandsRequiringSubcommands); contains && len(arguments) < 3 {
		return false, fmt.Sprintf(kArgumentErrorSubcommandRequired, command)
	}

	command = arguments[1]
	if contains, _ := SliceContainsString(command, kValidCommands); !contains {
		return false, fmt.Sprintf(kArgumentErrorUnknownCommand, command)
	}

	if len(arguments) > 2 {
		subcommand := arguments[2]
		contains, _ := SliceContainsString(subcommand, kValidSubcommands)
		if command != kCommandExclude &&
			command != kCommandInclude &&
			!contains {
			return false, fmt.Sprintf(kArgumentErrorUnknownSubcommand, subcommand)
		}
	}
	return true, ""
}

// EncodeJSON encodes an object to a JSON byte slice
func EncodeJSON(object interface{}) ([]byte, error) {
	return json.Marshal(object)
}

// DecodeJSON decodes a JSON byte slice into an object
func DecodeJSON(data []byte, object interface{}) error {
	return json.Unmarshal(data, object)
}

// Exclude excludes a directory from testing
func Exclude(directory string, config map[string]interface{}) {

	// If the directory isn't in the array, add it
	exclusions := config[kConfigKeyExclusions].([]string)
	if contains, _ := SliceContainsString(directory, exclusions); !contains || exclusions == nil {
		config[kConfigKeyExclusions] = append(exclusions, directory)
	}

	WriteConfig(config)

}

// Include includes a directory in testing
func Include(directory string, config map[string]interface{}) {

	// If the directory is in the array, remove it
	exclusions := config[kConfigKeyExclusions].([]string)
	if contains, index := SliceContainsString(directory, exclusions); contains {
		config[kConfigKeyExclusions] = append(exclusions[:index], exclusions[index+1:]...)
	}

	WriteConfig(config)
}

// ConfigEmpty determines if the configuration object is empty, allowing the configuration file to be deleted
func ConfigEmpty(config map[string]interface{}) bool {

	empty := false

	if len(config[kConfigKeyExclusions].([]string)) == 0 {
		empty = true
	}

	return empty

}

// WriteConfig writes the configuration to disk
func WriteConfig(config map[string]interface{}) {

	// The configuration is empty. Delete the file.
	if ConfigEmpty(config) {
		os.Remove(kConfigFilename)
	} else {

		data, error := EncodeJSON(config)

		if error != nil {
			fmt.Printf("\n%s\n\n", kErrorSavingFile)
		}

		error = ioutil.WriteFile(kConfigFilename, data, 0644)

		if error != nil {
			fmt.Printf("\n%s\n\n", kErrorSavingFile)
		}
	}

}

// FormatExclusionsForPrint returns a string detailing all excluded directories.
func FormatExclusionsForPrint(exclusions []string) string {

	excludedPackages := strings.Join(exclusions, "\n\t")
	return fmt.Sprintf("Excluded Directories:\n\t%s", excludedPackages)

}

// RunTests recurses all directories and installs tests dependencies, then runs tests
func RunTests(subcommand string, exclusions []string) (int, int) {
	directory, error := os.Getwd()
	if error != nil {
		fmt.Printf(kErrorCurrentDirectory, error)
	}

	// We have a subcommand
	if len(subcommand) != 0 {
		if subcommand == kSubcommandAll {
			// Pass nil as exclusions to run all tests
			return RecurseDirectories(directory, nil, kShellCommandInstallDependencies, kShellCommandRunTest)
		}
	}

	return RecurseDirectories(directory, exclusions, kShellCommandInstallDependencies, kShellCommandRunTest)
}

// InstallTestDependencies recurses all directories and installs test dependencies
func InstallTestDependencies() (int, int) {
	directory, error := os.Getwd()
	if error != nil {
		fmt.Printf(kErrorCurrentDirectory, error)
	}
	return RecurseDirectories(directory, nil, kShellCommandInstallDependencies)
}

// RecurseDirectories recurses all directories and runs the given commands
// exclusions contains directories to be skipped
// Multiple commands may be passed and each will be run in sequence
func RecurseDirectories(directory string, exclusions []string, commands ...string) (int, int) {

	testsRun := 0
	testsFailed := 0

	// If this directory is not contained in the exclusions slice
	dirSplit := strings.Split(directory, "/")
	dirName := dirSplit[len(dirSplit)-1]
	if success, _ := SliceContainsString(dirName, exclusions); !success {

		directoryHandle, error := os.Open(directory)
		if error != nil {
			fmt.Printf(kErrorRecursingDirectories, error)
			os.Exit(1)
		}
		files, error := directoryHandle.Readdir(-1)
		if error != nil {
			fmt.Printf(kErrorRecursingDirectories, error)
			os.Exit(1)
		}

		testFileDetected := false
		for _, file := range files {

			// If this is a directory, recurse into it
			if file.IsDir() {
				tempTestsRun, tempTestsFailed := RecurseDirectories(fmt.Sprintf("%s/%s", directory, file.Name()), exclusions, commands...)
				testsRun += tempTestsRun
				testsFailed += tempTestsFailed
			} else {
				// Determine if the filename contains "_test.go", indicating a testable directory
				if testFileDetected == false && strings.Contains(file.Name(), "_test.go") {
					testFileDetected = true
				}
			}

		}

		if testFileDetected {

			testsRun++

			succeeded := true

			for i := 0; i < len(commands); i++ {

				// Explode the test string and extract the command, arguments
				splitCommand := strings.Split(commands[i], " ")

				command := splitCommand[0]
				arguments := splitCommand[1:]

				shellCommand := exec.Command(command, arguments...)
				shellCommand.Dir = directory

				if output, error := shellCommand.Output(); error != nil {
					testsFailed++
					succeeded = false
					// Test failed, print the test output
					fmt.Printf("\n\n%s\n\n", output)
				}
			}
			if succeeded {
				// Print a . to indicate progress
				fmt.Print(".")
			}
		}
	}
	return testsRun, testsFailed

}

func main() {

	arguments := os.Args

	// Verify the arguments
	if success, details := VerifyArguments(arguments); !success {
		fmt.Printf("\n%s\n\n", details)
		os.Exit(1)
	}

	// Set up the basic configuration object in case we have no saved configuration file
	var config = make(map[string]interface{})
	config[kConfigKeyExclusions] = make([]string, 0)

	// If a configuration file exists, load and decode it		
	if fileData, fileError := ioutil.ReadFile(kConfigFilename); fileError == nil {
		if decodeError := DecodeJSON(fileData, &config); decodeError != nil {
			fmt.Printf("There was an error parsing your configuration file: %s\n\n", decodeError)
			os.Exit(1)
		} else {
			// Convert the []interface{} to []string to make life easier
			config[kConfigKeyExclusions] = StringSliceFromInterfaceSlice(config[kConfigKeyExclusions].([]interface{}))
		}
	}

	exclusions := config[kConfigKeyExclusions].([]string)
	command := arguments[1]
	subcommand := ""
	if len(arguments) == 3 {
		subcommand = arguments[2]
	}

	switch command {
	case kCommandRun:
		fmt.Printf("\nRunning tests")
		testsRun, testsFailed := RunTests(subcommand, exclusions)
		fmt.Printf("\n%d tests run. %d tests succeeded. %d tests failed. [%.0f%% success]\n\n", testsRun, testsRun-testsFailed, testsFailed, (float32((testsRun-testsFailed))/float32(testsRun))*100)
	case kCommandInstall:
		fmt.Printf("\nInstalling test dependencies")
		testsRun, testsFailed := InstallTestDependencies()
		fmt.Printf("\n%d installed. %d failed. [%.0f%% success]\n\n", testsRun-testsFailed, testsFailed, (float32((testsRun-testsFailed))/float32(testsRun))*100)
	case kCommandExclude:
		Exclude(subcommand, config)
		fmt.Printf("\nExcluded: %s\n\n", subcommand)
	case kCommandInclude:
		Include(subcommand, config)
		fmt.Printf("\nIncluded: %s\n\n", subcommand)
	case kCommandExclusions:
		excludedPackages := FormatExclusionsForPrint(exclusions)
		fmt.Printf("\n%s\n\n", excludedPackages)
	}

}
