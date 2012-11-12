package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// argumentErrorUsage is printed when no arguments are passed to the program.
var argumentErrorUsage string = `gort installs dependencies and/or runs tests, recursively, starting from the current working directory

Usage: gort command [subcommand]

Valid commands are:

	test - Runs tests
	test <directory name> - Runs tests in the named directory
	test all - Runs tests, including excluded directories
	install - Installs dependencies for tests
	install <directory name> - Installs dependencies for tests in the named directory
	install all - Installs dependencies for tests, including excluded directories
	vet - Vets code
	vet <directory name> - Vets code in the named directory
	vet all - Vets code, including excluded directories
	exclude <directory name> - Excludes a directory from recursion
	include <directory name> - Includes a directory in recursion
	exclusions - Prints a list of excluded directories

Version 1.0`

var (
	// argumentErrorUnknownCommand is printed when an unknown command is passed to the program.
	argumentErrorUnknownCommand string = "Unknown command: %s"

	// argumentErrorUnknownSubcommandCommand is printed when an unknown subcommand is passed to the program.
	argumentErrorUnknownSubcommand string = "Unknown subcommand: %s"

	// argumentErrorSubcommandRequired is printed when a subcommand is required, but was not passed.
	argumentErrorSubcommandRequired string = "%s requires a subcommand"

	// errorSavingFile is printed when an error occurs attempting to save the configuration file.
	errorSavingFile = "There was an error attempting to save your configuration file."

	// errorRecursingDirectories is printed when an error occurs recursing through the directory structure.
	errorRecursingDirectories = "There was an error when attempting to recurse directories: %s"

	// errorCurrentDirectory is printed when an error occurs attempting to get the current working directory.
	errorCurrentDirectory = "There was an error attempting to get directory in which gort is being run: %s"

	// commandTest is the string for the test command.
	commandTest string = "test"

	// commandVey is the string for the vet command.
	commandVet string = "vet"

	// commandHelp is the string for the help command.
	commandHelp string = "help"

	// commandInstall is the string for the install command.
	commandInstall string = "install"

	// commandExclude is the string for the exclude command.
	commandExclude string = "exclude"

	// commandInclude is the string for the include command.
	commandInclude string = "include"

	// commandExclusions is the string for the exclusions command.
	commandExclusions string = "exclusions"

	// validCommands contains the valid top level commands. Used to verify the top level command is sane.
	validCommands = []string{commandTest, commandVet, commandHelp, commandInstall, commandExclude, commandInclude, commandExclusions}

	// commandsRequiringSubcommands contains the top level commands that require a subcommand. Used to enforce subcommands when required.
	commandsRequiringSubcommands = []string{commandExclude, commandInclude}

	// subcommandAll is the string for the all subcommand. Used with the run command.
	subcommandAll string = "all"

	// validSubcommands contains the valid subcommands. Used to verify subcommand is sane.
	validSubcommands = []string{subcommandAll}

	// configKeyExclusions is the string for the key in the configuration object at which the exclusions list is stored
	configKeyExclusions = "exclusions"

	// configFilename is the string for the name of the gort configuration file
	configFilename = ".gort"

	// shellCommandInstallDependencies is the string for the shell command to run when installing dependencies
	shellCommandInstallDependencies string = "go test -i"

	// shellCommandTest is the string for the shell command to run when executing tests
	shellCommandTest string = "go test"

	// shellCommandVet is the string for the shell command to run when vetting code
	shellCommandVet string = "go vet"

	// fileTypeTest is the string contining the file type to look for when running/installing tests
	fileTypeTest string = "_test.go"

	// fileTypeGo is the string containing the file type to look for when vetting tests
	fileTypeVet string = ".go"
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
		return false, argumentErrorUsage
	}

	// Verify that the command is valid
	command := arguments[1]
	if contains, _ := SliceContainsString(command, commandsRequiringSubcommands); contains && len(arguments) < 3 {
		return false, fmt.Sprintf(argumentErrorSubcommandRequired, command)
	}

	command = arguments[1]
	if contains, _ := SliceContainsString(command, validCommands); !contains {
		return false, fmt.Sprintf(argumentErrorUnknownCommand, command)
	}

	if len(arguments) > 2 {
		subcommand := arguments[2]
		contains, _ := SliceContainsString(subcommand, validSubcommands)
		if command != commandExclude &&
			command != commandInclude &&
			command != commandTest &&
			command != commandInstall &&
			command != commandVet &&
			!contains {
			return false, fmt.Sprintf(argumentErrorUnknownSubcommand, subcommand)
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
	exclusions := config[configKeyExclusions].([]string)
	if contains, _ := SliceContainsString(directory, exclusions); !contains || exclusions == nil {
		config[configKeyExclusions] = append(exclusions, directory)
	}

	WriteConfig(config)

}

// Include includes a directory in testing
func Include(directory string, config map[string]interface{}) {

	// If the directory is in the array, remove it
	exclusions := config[configKeyExclusions].([]string)
	if contains, index := SliceContainsString(directory, exclusions); contains {
		config[configKeyExclusions] = append(exclusions[:index], exclusions[index+1:]...)
	}

	WriteConfig(config)
}

// ConfigEmpty determines if the configuration object is empty, allowing the configuration file to be deleted
func ConfigEmpty(config map[string]interface{}) bool {

	empty := false

	if len(config[configKeyExclusions].([]string)) == 0 {
		empty = true
	}

	return empty

}

// WriteConfig writes the configuration to disk
func WriteConfig(config map[string]interface{}) {

	// The configuration is empty. Delete the file.
	if ConfigEmpty(config) {
		os.Remove(configFilename)
	} else {

		data, error := EncodeJSON(config)

		if error != nil {
			fmt.Printf("\n%s\n\n", errorSavingFile)
		}

		error = ioutil.WriteFile(configFilename, data, 0644)

		if error != nil {
			fmt.Printf("\n%s\n\n", errorSavingFile)
		}
	}

}

// FormatExclusionsForPrint returns a string detailing all excluded directories.
func FormatExclusionsForPrint(exclusions []string) string {

	excludedPackages := strings.Join(exclusions, "\n\t")
	return fmt.Sprintf("Excluded Directories:\n\t%s", excludedPackages)

}

func installAndExecuteTests(subcommand string, exclusions []string) {

	fmt.Printf("\nInstalling test dependencies")
	// We have a subcommand
	var testsRun int
	var testsFailed int
	if len(subcommand) != 0 {
		if subcommand == subcommandAll {
			// Pass nil as exclusions to run all tests
			testsRun, testsFailed = InstallTestDependencies(nil, "")
		} else {
			// Assume a single test directory to be run
			testsRun, testsFailed = InstallTestDependencies(nil, subcommand)
		}
	} else {
		testsRun, testsFailed = InstallTestDependencies(exclusions, "")
	}

	fmt.Printf("\n%d installed. %d succeeded. %d failed. [%.0f%% success]\n\n", testsRun, testsRun-testsFailed, testsFailed, (float32((testsRun-testsFailed))/float32(testsRun))*100)

	fmt.Printf("\nRunning tests")
	testsRun, testsFailed = RunTests(subcommand, exclusions)
	fmt.Printf("\n%d run. %d succeeded. %d failed. [%.0f%% success]\n\n", testsRun, testsRun-testsFailed, testsFailed, (float32((testsRun-testsFailed))/float32(testsRun))*100)

}

// RunTests recurses all directories and installs tests dependencies, then runs tests
func RunTests(subcommand string, exclusions []string) (int, int) {
	directory, error := os.Getwd()
	if error != nil {
		fmt.Printf(errorCurrentDirectory, error)
	}

	// We have a subcommand
	if len(subcommand) != 0 {
		if subcommand == subcommandAll {
			// Pass nil as exclusions to run all tests
			return RecurseDirectories(directory, nil, "", shellCommandTest, fileTypeTest)
		} else {
			// Assume the subcommand is a directory to be tested
			return RecurseDirectories(directory, nil, subcommand, shellCommandTest, fileTypeTest)
		}
	}

	return RecurseDirectories(directory, exclusions, "", shellCommandTest, fileTypeTest)
}

// InstallTestDependencies recurses all directories and installs test dependencies
func InstallTestDependencies(exclusions []string, targetDirectory string) (int, int) {
	directory, error := os.Getwd()
	if error != nil {
		fmt.Printf(errorCurrentDirectory, error)
	}
	// We have a targetDirectory
	if len(targetDirectory) != 0 {
		if targetDirectory == subcommandAll {
			return RecurseDirectories(directory, nil, "", shellCommandInstallDependencies, fileTypeTest)
		} else {
			return RecurseDirectories(directory, nil, targetDirectory, shellCommandInstallDependencies, fileTypeTest)
		}
	}
	return RecurseDirectories(directory, exclusions, targetDirectory, shellCommandInstallDependencies, fileTypeTest)
}

// VetPackages recurses all directories and vets packages
func VetPackages(exclusions []string, targetDirectory string) (int, int) {
	directory, error := os.Getwd()
	if error != nil {
		fmt.Printf(errorCurrentDirectory, error)
	}
	// We have a targetDirectory
	if len(targetDirectory) != 0 {
		if targetDirectory == subcommandAll {
			return RecurseDirectories(directory, nil, "", shellCommandVet, fileTypeVet)
		} else {
			return RecurseDirectories(directory, nil, targetDirectory, shellCommandVet, fileTypeVet)
		}
	}

	return RecurseDirectories(directory, exclusions, targetDirectory, shellCommandVet, fileTypeVet)
}

// RecurseDirectories recurses all directories and runs the given commands
// exclusions contains directories to be skipped
// Multiple commands may be passed and each will be run in sequence
func RecurseDirectories(directory string, exclusions []string, targetDirectory, command string, fileType string) (int, int) {

	testsRun := 0
	testsFailed := 0

	// If this directory is not contained in the exclusions slice
	dirSplit := strings.Split(directory, "/")
	dirName := dirSplit[len(dirSplit)-1]
	if success, _ := SliceContainsString(dirName, exclusions); !success {

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

		directoryParts := strings.Split(directory, string(os.PathSeparator))
		directoryName := directoryParts[len(directoryParts)-1]
		testFileDetected := false
		for _, file := range files {

			// If this is a directory, recurse into it, unless we are already in our targetDirectory			
			if file.IsDir() && directoryName != targetDirectory {
				tempTestsRun, tempTestsFailed := RecurseDirectories(fmt.Sprintf("%s/%s", directory, file.Name()), exclusions, targetDirectory, command, fileTypeTest)
				testsRun += tempTestsRun
				testsFailed += tempTestsFailed
			} else {
				// Determine if the filename contains "_test.go", indicating a testable directory
				if testFileDetected == false && strings.Contains(file.Name(), fileType) {
					testFileDetected = true
				}
			}

		}

		if (testFileDetected && len(targetDirectory) == 0) || (testFileDetected && directoryName == targetDirectory) {
			testsRun++

			succeeded := true

			// Explode the test string and extract the command, arguments
			splitCommand := strings.Split(command, " ")

			command := splitCommand[0]
			arguments := splitCommand[1:]

			shellCommand := exec.Command(command, arguments...)
			shellCommand.Dir = directory

			if output, error := shellCommand.CombinedOutput(); error != nil {

				testsFailed++
				succeeded = false

				// Test failed, print the test output
				fmt.Printf("\n\n%s", output)
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

	// Set up the basic configuration object in case we have no saved configuration file
	var config = make(map[string]interface{})
	config[configKeyExclusions] = make([]string, 0)

	// If a configuration file exists, load and decode it		
	if fileData, fileError := ioutil.ReadFile(configFilename); fileError == nil {
		if decodeError := DecodeJSON(fileData, &config); decodeError != nil {
			fmt.Printf("There was an error parsing your configuration file: %s\n\n", decodeError)
			os.Exit(1)
		} else {
			// Convert the []interface{} to []string to make life easier
			config[configKeyExclusions] = StringSliceFromInterfaceSlice(config[configKeyExclusions].([]interface{}))
		}
	}

	exclusions := config[configKeyExclusions].([]string)

	if len(arguments) == 1 {
		installAndExecuteTests("", exclusions)
		os.Exit(0)
	}

	// Verify the arguments
	if success, details := VerifyArguments(arguments); !success {
		fmt.Printf("\n%s\n\n", details)
		os.Exit(1)
	}

	command := arguments[1]
	subcommand := ""
	if len(arguments) == 3 {
		subcommand = arguments[2]
	}

	switch command {
	case commandTest:
		installAndExecuteTests(subcommand, exclusions)
	case commandVet:
		fmt.Printf("\nVetting packages")
		vetsRun, vetsFailed := VetPackages(exclusions, subcommand)
		fmt.Printf("\n%d vetted. %d passed. %d failed. [%.0f%% passed]\n\n", vetsRun, vetsRun-vetsFailed, vetsFailed, (float32((vetsRun-vetsFailed))/float32(vetsRun))*100)
	case commandHelp:
		fmt.Printf("\n%s\n\n", argumentErrorUsage)
	case commandInstall:
		fmt.Printf("\nInstalling test dependencies")
		testsRun, testsFailed := InstallTestDependencies(exclusions, subcommand)
		fmt.Printf("\n%d installed. %d succeeded. %d failed. [%.0f%% success]\n\n", testsRun, testsRun-testsFailed, testsFailed, (float32((testsRun-testsFailed))/float32(testsRun))*100)
	case commandExclude:
		Exclude(subcommand, config)
		fmt.Printf("\nExcluded: %s\n\n", subcommand)
	case commandInclude:
		Include(subcommand, config)
		fmt.Printf("\nIncluded: %s\n\n", subcommand)
	case commandExclusions:
		excludedPackages := FormatExclusionsForPrint(exclusions)
		fmt.Printf("\n%s\n\n", excludedPackages)
	}

}
