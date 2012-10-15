package main

import (
	"fmt"
	"github.com/stretchrcom/affirm/assert"
	"testing"
)

func TestVerifyArguments(t *testing.T) {

	// Test bad input

	arguments := []string{"professor"}
	success, details := VerifyArguments(arguments)

	if assert.False(t, success, details) {
		assert.Equal(t, details, kArgumentErrorUsage)
	}

	arguments = []string{"professor", "gorram", "browncoat", "harlot"}
	success, details = VerifyArguments(arguments)

	if assert.False(t, success, details) {
		assert.Equal(t, details, kArgumentErrorUsage)
	}

	arguments = []string{"professor", "gorram"}
	success, details = VerifyArguments(arguments)

	if assert.False(t, success, details) {
		assert.Equal(t, details, fmt.Sprintf(kArgumentErrorUnknownCommand, arguments[1]))
	}

	arguments = []string{"professor", "run", "fromTheLaw"}
	success, details = VerifyArguments(arguments)

	if assert.False(t, success, details) {
		assert.Equal(t, details, fmt.Sprintf(kArgumentErrorUnknownSubcommand, arguments[2]))
	}

	arguments = []string{"professor", "run", "fromTheLaw"}
	success, details = VerifyArguments(arguments)

	if assert.False(t, success, details) {
		assert.Equal(t, details, fmt.Sprintf(kArgumentErrorUnknownSubcommand, arguments[2]))
	}

	arguments = []string{"professor", "exclude"}
	success, details = VerifyArguments(arguments)

	if assert.False(t, success, details) {
		assert.Equal(t, details, fmt.Sprintf(kArgumentSubcommandRequired, arguments[1]))
	}

	arguments = []string{"professor", "include"}
	success, details = VerifyArguments(arguments)

	if assert.False(t, success, details) {
		assert.Equal(t, details, fmt.Sprintf(kArgumentSubcommandRequired, arguments[1]))
	}

	// Test good input

	arguments = []string{"professor", "run"}
	success, details = VerifyArguments(arguments)
	assert.True(t, success, details)

	arguments = []string{"professor", "run", "all"}
	success, details = VerifyArguments(arguments)
	assert.True(t, success, details)

	arguments = []string{"professor", "install"}
	success, details = VerifyArguments(arguments)
	assert.True(t, success, details)

	arguments = []string{"professor", "exclude", "package"}
	success, details = VerifyArguments(arguments)
	assert.True(t, success, details)

	arguments = []string{"professor", "include", "package"}
	success, details = VerifyArguments(arguments)
	assert.True(t, success, details)

	arguments = []string{"professor", "exclusions"}
	success, details = VerifyArguments(arguments)
	assert.True(t, success, details)

}

func TestEncodeJSON(t *testing.T) {

	var object = make(map[string]interface{})
	object[kConfigKeyExclusions] = []string{"alliance", "badger"}

	json, error := EncodeJSON(object)

	if assert.Nil(t, error) {
		assert.Equal(t, string(json), `{"exclusions":["alliance","badger"]}`)
	}

}

func TestDecodeJSON(t *testing.T) {

	var object = make(map[string]interface{})
	object[kConfigKeyExclusions] = []string{"alliance", "badger"}

	var decodedObject map[string]interface{}

	error := DecodeJSON([]byte(`{"exclusions":["alliance","badger"]}`), &decodedObject)

	if assert.Nil(t, error) {
		//TODO: fix assert.Equal to make this work
		//assert.Equal(t, object, decodedObject)
	}

	error = DecodeJSON([]byte(`whee{}{[[;;:`), &decodedObject)

	assert.NotNil(t, error)
}

func TestExclude(t *testing.T) {

	var config = make(map[string]interface{})
	config[kConfigKeyExclusions] = make([]string, 0)

	Exclude("badger", config)

	assert.Equal(t, 1, len(config[kConfigKeyExclusions].([]string)))
	assert.Equal(t, config[kConfigKeyExclusions].([]string)[0], "badger")

}

func TestInclude(t *testing.T) {

	var config = make(map[string]interface{})
	config[kConfigKeyExclusions] = make([]string, 0)

	Exclude("badger", config)
	Exclude("reavers", config)

	assert.Equal(t, 2, len(config[kConfigKeyExclusions].([]string)), fmt.Sprintf("%v", config[kConfigKeyExclusions].([]string)))

	Include("reavers", config)

	assert.Equal(t, 1, len(config[kConfigKeyExclusions].([]string)), fmt.Sprintf("%v", config[kConfigKeyExclusions].([]string)))

	Include("badger", config)

	assert.Equal(t, 0, len(config[kConfigKeyExclusions].([]string)), fmt.Sprintf("%v", config[kConfigKeyExclusions].([]string)))

}

func TestFormatExclusionsForPrint(t *testing.T) {

	var config = make(map[string]interface{})
	config[kConfigKeyExclusions] = make([]string, 0)

	Exclude("badger", config)
	Exclude("reavers", config)

	exclusionString := "Excluded Directories:\n\tbadger\n\treavers"

	assert.Equal(t, FormatExclusionsForPrint(config[kConfigKeyExclusions].([]string)), exclusionString)

	Include("badger", config)
	Include("reavers", config)

}
