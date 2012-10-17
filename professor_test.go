package main

import (
	"fmt"
	"github.com/stretchrcom/testify/assert"
	"testing"
)

func TestVerifyArguments(t *testing.T) {

	// Test bad input

	arguments := []string{"gort"}
	success, details := VerifyArguments(arguments)

	if assert.False(t, success, details) {
		assert.Equal(t, details, argumentErrorUsage)
	}

	arguments = []string{"gort", "gorram", "browncoat", "harlot"}
	success, details = VerifyArguments(arguments)

	if assert.False(t, success, details) {
		assert.Equal(t, details, argumentErrorUsage)
	}

	arguments = []string{"gort", "gorram"}
	success, details = VerifyArguments(arguments)

	if assert.False(t, success, details) {
		assert.Equal(t, details, fmt.Sprintf(argumentErrorUnknownCommand, arguments[1]))
	}

	arguments = []string{"gort", "test", "fromTheLaw"}
	success, details = VerifyArguments(arguments)

	if assert.False(t, success, details) {
		assert.Equal(t, details, fmt.Sprintf(argumentErrorUnknownSubcommand, arguments[2]))
	}

	arguments = []string{"gort", "test", "fromTheLaw"}
	success, details = VerifyArguments(arguments)

	if assert.False(t, success, details) {
		assert.Equal(t, details, fmt.Sprintf(argumentErrorUnknownSubcommand, arguments[2]))
	}

	arguments = []string{"gort", "exclude"}
	success, details = VerifyArguments(arguments)

	if assert.False(t, success, details) {
		assert.Equal(t, details, fmt.Sprintf(argumentErrorSubcommandRequired, arguments[1]))
	}

	arguments = []string{"gort", "include"}
	success, details = VerifyArguments(arguments)

	if assert.False(t, success, details) {
		assert.Equal(t, details, fmt.Sprintf(argumentErrorSubcommandRequired, arguments[1]))
	}

	// Test good input

	arguments = []string{"gort", "test"}
	success, details = VerifyArguments(arguments)
	assert.True(t, success, details)

	arguments = []string{"gort", "test", "all"}
	success, details = VerifyArguments(arguments)
	assert.True(t, success, details)

	arguments = []string{"gort", "install"}
	success, details = VerifyArguments(arguments)
	assert.True(t, success, details)

	arguments = []string{"gort", "exclude", "package"}
	success, details = VerifyArguments(arguments)
	assert.True(t, success, details)

	arguments = []string{"gort", "include", "package"}
	success, details = VerifyArguments(arguments)
	assert.True(t, success, details)

	arguments = []string{"gort", "exclusions"}
	success, details = VerifyArguments(arguments)
	assert.True(t, success, details)

}

func TestEncodeJSON(t *testing.T) {

	var object = make(map[string]interface{})
	object[configKeyExclusions] = []string{"alliance", "badger"}

	json, error := EncodeJSON(object)

	if assert.Nil(t, error) {
		assert.Equal(t, string(json), `{"exclusions":["alliance","badger"]}`)
	}

}

func TestDecodeJSON(t *testing.T) {

	var object = make(map[string]interface{})
	object[configKeyExclusions] = []interface{}{"alliance", "badger"}

	var decodedObject map[string]interface{}

	error := DecodeJSON([]byte(`{"exclusions":["alliance","badger"]}`), &decodedObject)

	if assert.Nil(t, error) {
		assert.Equal(t, object, decodedObject)
	}

	error = DecodeJSON([]byte(`whee{}{[[;;:`), &decodedObject)

	assert.NotNil(t, error)
}

func TestExclude(t *testing.T) {

	var config = make(map[string]interface{})
	config[configKeyExclusions] = make([]string, 0)

	Exclude("badger", config)

	assert.Equal(t, 1, len(config[configKeyExclusions].([]string)))
	assert.Equal(t, config[configKeyExclusions].([]string)[0], "badger")

}

func TestInclude(t *testing.T) {

	var config = make(map[string]interface{})
	config[configKeyExclusions] = make([]string, 0)

	Exclude("badger", config)
	Exclude("reavers", config)

	assert.Equal(t, 2, len(config[configKeyExclusions].([]string)), fmt.Sprintf("%v", config[configKeyExclusions].([]string)))

	Include("reavers", config)

	assert.Equal(t, 1, len(config[configKeyExclusions].([]string)), fmt.Sprintf("%v", config[configKeyExclusions].([]string)))

	Include("badger", config)

	assert.Equal(t, 0, len(config[configKeyExclusions].([]string)), fmt.Sprintf("%v", config[configKeyExclusions].([]string)))

}

func TestFormatExclusionsForPrint(t *testing.T) {

	var config = make(map[string]interface{})
	config[configKeyExclusions] = make([]string, 0)

	Exclude("badger", config)
	Exclude("reavers", config)

	exclusionString := "Excluded Directories:\n\tbadger\n\treavers"

	assert.Equal(t, FormatExclusionsForPrint(config[configKeyExclusions].([]string)), exclusionString)

	Include("badger", config)
	Include("reavers", config)

}
