package main

import (
	"fmt"
)

var kArgumentErrorUsage string = `usage`
var kArgumentErrorUnknownCommand string = "Unknown command: %s"
var kArgumentErrorUnknownSubcommand string = "Unknown subcommand: %s"

var kCommandRun string = "run"
var kCommandInstall string = "install"
var kCommandExclude string = "exclude"
var kCommandInclude string = "include"
var kCommandExclusions string = "exclusions"
var kValidCommands = []string{kCommandRun, kCommandInstall, kCommandExclude, kCommandInclude, kCommandExclusions}

var kSubcommandExcluded string = "excluded"
var kSubcommandAll string = "all"
var kValidSubcommands = []string{kSubcommandExcluded, kSubcommandAll}

// Returns true if the string slice contains a given string
func SliceContainsString(target string, slice []string) bool {
	for _, value := range slice {
		if value == target {
			return true
		}
	}
	return false
}

// Verifies that the arguments passed are sane and returns the 
// If the arguments are not sane, returns false and a string detailing the proper usage.
func VerifyArguments(arguments []string) (bool, string) {

	if len(arguments) == 1 || len(arguments) > 3 {
		return false, kArgumentErrorUsage
	}

	// Verify that the command is valid
	command := arguments[1]
	if !SliceContainsString(command, kValidCommands) {
		return false, fmt.Sprintf(kArgumentErrorUnknownCommand, command)
	}

	if len(arguments) > 2 {
		subcommand := arguments[2]
		if command != kCommandExclude &&
			command != kCommandInclude &&
			!SliceContainsString(subcommand, kValidSubcommands) {
			return false, fmt.Sprintf(kArgumentErrorUnknownSubcommand, subcommand)
		}
	}
	return true, ""
}
