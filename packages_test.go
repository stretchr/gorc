package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPackagesFilterPackages(t *testing.T) {

	packages := []string{
		"github.com/stretchr/testify/assert",
		"github.com/stretchr/testify/mock",
		"github.com/stretchr/testify/assert",
		"github.com/stretchr/goweb/",
		"github.com/stretchr/goweb/helpers",
	}

	filteredPackages := filterPackages(packages, "assert", []string{})
	if assert.Equal(t, len(filteredPackages), 1) {
		assert.Equal(t, filteredPackages[0], "github.com/stretchr/testify/assert")
	}

	filteredPackages = filterPackages(packages, "all", []string{"testify"})
	assert.Equal(t, len(filteredPackages), 5)

	filteredPackages = filterPackages(packages, "", []string{"testify"})
	if assert.Equal(t, len(filteredPackages), 2) {
		assert.Equal(t, filteredPackages[0], "github.com/stretchr/gowebs/")
		assert.Equal(t, filteredPackages[1], "github.com/stretchr/goweb/helpers")
	}

}
