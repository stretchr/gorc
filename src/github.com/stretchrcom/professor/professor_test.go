package main

import (
	"fmt"
	"github.com/stretchrcom/affirm/assert"
	"testing"
)

func TestVerifyArguments(t *testing.T) {

	arguments := []string{"professor"}
	success, details := VerifyArguments(arguments)

	// Test bad input

	if assert.False(t, success) {
		assert.Equal(t, details, kArgumentErrorUsage)
	}

	arguments = []string{"professor", "gorram", "browncoat", "harlot"}
	success, details = VerifyArguments(arguments)

	if assert.False(t, success) {
		assert.Equal(t, details, kArgumentErrorUsage)
	}

	arguments = []string{"professor", "gorram"}
	success, details = VerifyArguments(arguments)

	if assert.False(t, success) {
		assert.Equal(t, details, fmt.Sprintf(kArgumentErrorUnknownCommand, arguments[1]))
	}

	arguments = []string{"professor", "run", "fromTheLaw"}
	success, details = VerifyArguments(arguments)

	if assert.False(t, success) {
		assert.Equal(t, details, fmt.Sprintf(kArgumentErrorUnknownSubcommand, arguments[2]))
	}

	// Test good input

	arguments = []string{"professor", "run"}
	success, details = VerifyArguments(arguments)
	assert.True(t, success, details)

	arguments = []string{"professor", "run", "excluded"}
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
