package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// TODO: print useful information for the usage
var kArgumentErrorUsage string = `usage`
var kArgumentErrorUnknownCommand string = "Unknown command: %s"
var kArgumentErrorUnknownSubcommand string = "Unknown subcommand: %s"

var kErrorSavingFile = "There was an error attempting to save your configuration file."

var kCommandRun string = "run"
var kCommandInstall string = "install"
var kCommandExclude string = "exclude"
var kCommandInclude string = "include"
var kCommandExclusions string = "exclusions"
var kValidCommands = []string{kCommandRun, kCommandInstall, kCommandExclude, kCommandInclude, kCommandExclusions}

var kSubcommandExcluded string = "excluded"
var kSubcommandAll string = "all"
var kValidSubcommands = []string{kSubcommandExcluded, kSubcommandAll}

var kConfigKeyExclusions = "exclusions"
var kConfigFilename = ".professor"

// Returns true if the string slice contains a given string
func SliceContainsString(target string, slice []string) (bool, int) {
	for index, value := range slice {
		if value == target {
			return true, index
		}
	}
	return false, -1
}

// Verifies that the arguments passed are sane and returns the 
// If the arguments are not sane, returns false and a string detailing the proper usage.
func VerifyArguments(arguments []string) (bool, string) {

	if len(arguments) == 1 || len(arguments) > 3 {
		return false, kArgumentErrorUsage
	}

	// Verify that the command is valid
	command := arguments[1]
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

	// Make sure we have an object to work with
	if config[kConfigKeyExclusions] == nil {
		config[kConfigKeyExclusions] = make([]string, 0)
	}

	// If the directory isn't in the array, add it
	exclusions := config[kConfigKeyExclusions].([]string)
	if contains, _ := SliceContainsString(directory, exclusions); !contains || exclusions == nil {
		config[kConfigKeyExclusions] = append(exclusions, directory)
	}

	WriteConfig(config)

}

// Includes a directory in testing
func Include(directory string, config map[string]interface{}) {

	// Make sure we have an object to work with
	if config[kConfigKeyExclusions] == nil {
		config[kConfigKeyExclusions] = make([]string, 0)
	}

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

func main() {

	// Verify the arguments
	if success, details := VerifyArguments(os.Args); !success {
		fmt.Printf("\n%s\n\n", details)
	}

	var config = make(map[string]interface{})

	// If a configuration file exists, load and decode it		
	if fileData, fileError := ioutil.ReadFile(kConfigFilename); fileError == nil {
		if decodeError := DecodeJSON(fileData, &config); decodeError != nil {
			fmt.Printf("There was an error parsing your configuration file: %s\n\n", decodeError)
			os.Exit(1)
		}
	}

}
