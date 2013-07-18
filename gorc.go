package main

import (
	"fmt"
	"github.com/stretchr/commander"
	"github.com/stretchr/stew/objects"
	"os"
)

func getwd() string {
	directory, error := os.Getwd()
	if error != nil {
		fmt.Printf(errorCurrentDirectory, error)
		os.Exit(1)
	}
	return directory
}

var exclusions []string

func main() {

	var config = readConfig()
	exclusions = config[configKeyExclusions].([]string)

	commander.Go(func() {
		commander.Map(commander.DefaultCommand, "", "",
			func(args objects.Map) {
				//name := args.GetString("name")
			})

		commander.Map("test [name=(string)]", "Runs tests, or named test",
			"If no name argument is specified, runs all tests recursively. If a name argument is specified, runs just that test, unless the argument is \"all\", in which case it runs all tests, including those in the exclusion list.",
			func(args objects.Map) {
				//name := args.GetString("name")
				fmt.Printf("Found packages: %v\n", buildPackageList(getwd(), fileTypeTest))
			})

		commander.Map("install [name=(string)]", "Installs tests, or named test",
			"If no name argument is specified, installs all tests recursively. If a name argument is specified, installs just that test, unless the argument is \"all\", in which case it installs all tests, including those in the exclusion list.",
			func(args objects.Map) {
				//name := args.GetString("name")
			})

		commander.Map("vet [name=(string)]", "Vets packages, or named package",
			"If no name argument is specified, vets all packages recursively. If a name argument is specified, vets just that package, unless the argument is \"all\", in which case it vets all packages, including those in the exclusion list.",
			func(args objects.Map) {
				//name := args.GetString("name")
			})

		commander.Map("race [name=(string)]", "Runs race detector on tests, or named test",
			"If no name argument is specified, race tests all tests recursively. If a name argument is specified, vets just that test, unless the argument is \"all\", in which case it vets all tests, including those in the exclusion list.",
			func(args objects.Map) {
				//name := args.GetString("name")
			})

		commander.Map("exclude name=(string)", "Excludes the named directory from recursion",
			"An excluded directory will be skipped when walking the directory tree. Any subdirectories of the excluded directory will also be skipped.",
			func(args objects.Map) {
				name := args.GetString("name")
				exclude(name, config)
				fmt.Printf("\nExcluded \"%s\" from being examined during recursion.\n", name)
				config = readConfig()
				exclusions = config[configKeyExclusions].([]string)
				fmt.Printf("\n%s\n\n", formatExclusionsForPrint(exclusions))
			})

		commander.Map("include name=(string)", "Removes the named directory from the exclusion list", "",
			func(args objects.Map) {
				name := args.GetString("name")
				include(name, config)
				fmt.Printf("\nRemoved \"%s\" from the exclusion list.\n", name)
				fmt.Printf("\n%s\n\n", formatExclusionsForPrint(exclusions))
			})

		commander.Map("exclusions", "Prints the exclusion list", "",
			func(args objects.Map) {
				fmt.Printf("\n%s\n\n", formatExclusionsForPrint(exclusions))
			})
	})

}
