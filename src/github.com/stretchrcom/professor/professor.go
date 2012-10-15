package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// TODO: print useful information for the usage
var kArgumentErrorUsage string = `Usage: professor command [subcommand]

Valid commands are:

	run - Runs tests, recursively, starting from this directory
	run [all] - Runs tests, recursively, including excluded directories, starting from this directory
	install - Installs dependencies, recursively, for tests
	exclude <directory name> - Excludes a directory from recursion
	include <directory name> - Includes a directory in recursion
	exclusions - Prints a list of excluded directories`

var kArgumentErrorUnknownCommand string = "Unknown command: %s"
var kArgumentErrorUnknownSubcommand string = "Unknown subcommand: %s"
var kArgumentSubcommandRequired string = "%s requires a subcommand"

var kErrorSavingFile = "There was an error attempting to save your configuration file."
var kErrorRecursingDirectories = "There was an error when attempting to recurse directories: %s"
var kErrorCurrentDirectory = "There was an error attempting to get directory in which professor is being run: %s"

var kCommandRun string = "run"
var kCommandInstall string = "install"
var kCommandExclude string = "exclude"
var kCommandInclude string = "include"
var kCommandExclusions string = "exclusions"
var kValidCommands = []string{kCommandRun, kCommandInstall, kCommandExclude, kCommandInclude, kCommandExclusions}
var kCommandsRequiringSubcommands = []string{kCommandExclude, kCommandInclude}

var kSubcommandAll string = "all"
var kValidSubcommands = []string{kSubcommandAll}

var kConfigKeyExclusions = "exclusions"
var kConfigFilename = ".professor"

var kShellCommandInstallDependencies string = "go test -i"
var kShellCommandRunTest string = "go test"

// Returns true if the string slice contains a given string
func SliceContainsString(target string, slice []string) (bool, int) {
	for index, value := range slice {
		if value == target {
			return true, index
		}
	}
	return false, -1
}

// Returns a []string from a []interface{}
func StringSliceFromInterfaceSlice(interfaceSlice []interface{}) []string {
	retval := make([]string, len(interfaceSlice))
	for i, str := range interfaceSlice {
		retval[i] = str.(string)
	}
	return retval
}

// Verifies that the arguments passed are sane and returns the 
// If the arguments are not sane, returns false and a string detailing the proper usage.
func VerifyArguments(arguments []string) (bool, string) {

	if len(arguments) == 1 || len(arguments) > 3 {
		return false, kArgumentErrorUsage
	}

	// Verify that the command is valid
	command := arguments[1]
	if contains, _ := SliceContainsString(command, kCommandsRequiringSubcommands); contains && len(arguments) < 3 {
		return false, fmt.Sprintf(kArgumentSubcommandRequired, command)
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

// Encodes an object to a JSON byte slice
func EncodeJSON(object interface{}) ([]byte, error) {
	return json.Marshal(object)
}

// Decodes a JSON byte slice into an object
func DecodeJSON(data []byte, object interface{}) error {
	return json.Unmarshal(data, object)
}

// Excludes a directory from testing
func Exclude(directory string, config map[string]interface{}) {

	// If the directory isn't in the array, add it
	exclusions := config[kConfigKeyExclusions].([]string)
	if contains, _ := SliceContainsString(directory, exclusions); !contains || exclusions == nil {
		config[kConfigKeyExclusions] = append(exclusions, directory)
	}

	WriteConfig(config)

}

// Includes a directory in testing
func Include(directory string, config map[string]interface{}) {

	// If the directory is in the array, remove it
	exclusions := config[kConfigKeyExclusions].([]string)
	if contains, index := SliceContainsString(directory, exclusions); contains {
		config[kConfigKeyExclusions] = append(exclusions[:index], exclusions[index+1:]...)
	}

	WriteConfig(config)
}

func ConfigEmpty(config map[string]interface{}) bool {

	empty := false

	if len(config[kConfigKeyExclusions].([]string)) == 0 {
		empty = true
	}

	return empty

}

// Writes the configuration to disk
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

// Returns a string detailing all excluded directories.
func FormatExclusionsForPrint(exclusions []string) string {

	excludedPackages := strings.Join(exclusions, "\n\t")
	return fmt.Sprintf("Excluded Directories:\n\t%s", excludedPackages)

}

// Recurses all directories and installes tests dependencies, then runs tests
func RunTests(subcommand string, exclusions []string) (int, int) {
	directory, error := os.Getwd()
	if error != nil {
		fmt.Printf(kErrorCurrentDirectory, error)
	}

	// We have a subcommand
	if len(subcommand) != 0 {
		if subcommand == kSubcommandAll {
			return RecurseDirectories(directory, nil, kShellCommandInstallDependencies, kShellCommandRunTest)
		}
	}

	return RecurseDirectories(directory, exclusions, kShellCommandInstallDependencies, kShellCommandRunTest)
}

// Recurses all directories and installs test dependencies
func InstallTestDependencies() (int, int) {
	directory, error := os.Getwd()
	if error != nil {
		fmt.Printf(kErrorCurrentDirectory, error)
	}
	return RecurseDirectories(directory, nil, kShellCommandInstallDependencies)
}

// Recurses all directories and runs commands
// exclusions contains directories to be skipped
// Multiple commands may be passed and each will be run
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

			if file.IsDir() {
				tempTestsRun, tempTestsFailed := RecurseDirectories(fmt.Sprintf("%s/%s", directory, file.Name()), exclusions, commands...)
				testsRun += tempTestsRun
				testsFailed += tempTestsFailed
			} else {
				if testFileDetected == false && strings.Contains(file.Name(), "_test.go") {
					testFileDetected = true
				}
			}

		}

		if testFileDetected {

			testsRun++
			fmt.Print(".")
			for i := 0; i < len(commands); i++ {

				splitCommand := strings.Split(commands[i], " ")

				command := splitCommand[0]
				arguments := splitCommand[1:]

				shellCommand := exec.Command(command, arguments...)
				shellCommand.Dir = directory

				if output, error := shellCommand.Output(); error != nil {
					testsFailed++
					fmt.Printf("\n\n%s\n\n", output)
				}
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

	var config = make(map[string]interface{})
	config[kConfigKeyExclusions] = make([]string, 0)

	// If a configuration file exists, load and decode it		
	if fileData, fileError := ioutil.ReadFile(kConfigFilename); fileError == nil {
		if decodeError := DecodeJSON(fileData, &config); decodeError != nil {
			fmt.Printf("There was an error parsing your configuration file: %s\n\n", decodeError)
			os.Exit(1)
		} else {
			// Convert various bits of data to appropriate types
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
