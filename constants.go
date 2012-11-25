package main

const (
	// configKeyExclusions is the string for the key in the configuration object at which the exclusions list is stored
	configKeyExclusions = "exclusions"

	// configFilename is the string for the name of the gort configuration file
	configFilename = ".gort"
)

const (
	// errorSavingFile is printed when an error occurs attempting to save the configuration file.
	errorSavingFile = "There was an error attempting to save your configuration file."

	// errorRecursingDirectories is printed when an error occurs recursing through the directory structure.
	errorRecursingDirectories = "There was an error when attempting to recurse directories: %s"

	// errorCurrentDirectory is printed when an error occurs attempting to get the current working directory.
	errorCurrentDirectory = "There was an error attempting to get directory in which gort is being run: %s"
)
