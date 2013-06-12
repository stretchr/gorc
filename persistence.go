package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// encodeJSON encodes an object to a JSON byte slice
func encodeJSON(object interface{}) ([]byte, error) {
	return json.Marshal(object)
}

// decodeJSON decodes a JSON byte slice into an object
func decodeJSON(data []byte, object interface{}) error {
	return json.Unmarshal(data, object)
}

// exclude excludes a directory from testing
func exclude(directory string, config map[string]interface{}) {

	// If the directory isn't in the array, add it
	exclusions := config[configKeyExclusions].([]string)
	if contains, _ := sliceContainsString(directory, exclusions); !contains || exclusions == nil {
		config[configKeyExclusions] = append(exclusions, directory)
	}

	writeConfig(config)

}

// Include includes a directory in testing
func include(directory string, config map[string]interface{}) {

	// If the directory is in the array, remove it
	exclusions := config[configKeyExclusions].([]string)
	if contains, index := sliceContainsString(directory, exclusions); contains {
		config[configKeyExclusions] = append(exclusions[:index], exclusions[index+1:]...)
	}

	writeConfig(config)
}

// configEmpty determines if the configuration object is empty, allowing the configuration file to be deleted
func configEmpty(config map[string]interface{}) bool {

	empty := false

	if len(config[configKeyExclusions].([]string)) == 0 {
		empty = true
	}

	return empty

}

// writeConfig writes the configuration to disk
func writeConfig(config map[string]interface{}) {

	// The configuration is empty. Delete the file.
	if configEmpty(config) {
		os.Remove(configFilename)
	} else {

		data, error := encodeJSON(config)

		if error != nil {
			fmt.Printf("\n%s\n\n", errorSavingFile)
		}

		error = ioutil.WriteFile(configFilename, data, 0644)

		if error != nil {
			fmt.Printf("\n%s\n\n", errorSavingFile)
		}
	}

}

// readConfig reads the configuration from disk
func readConfig() map[string]interface{} {
	var config = make(map[string]interface{})
	config[configKeyExclusions] = make([]string, 0)

	// If a configuration file exists, load and decode it
	if fileData, fileError := ioutil.ReadFile(configFilename); fileError == nil {
		if decodeError := decodeJSON(fileData, &config); decodeError != nil {
			fmt.Printf("There was an error parsing your configuration file: %s\n\n", decodeError)
			os.Exit(1)
		} else {
			// Convert the []interface{} to []string to make life easier
			config[configKeyExclusions] = stringSliceFromInterfaceSlice(config[configKeyExclusions].([]interface{}))
		}
	}
	return config
}
