package main

import (
	"strings"
	"strconv"
	"fmt"
	"github.com/stretchr/commander"
	"github.com/stretchr/objx"
	"os"
	"os/exec"
)

const (
	CommandInstallTests = iota
	CommandTest
	CommandInstall
	CommandRace
	CommandVet
	CommandCover
)

const (
	ResponseTypePass = iota
	ResponseTypeFail
)

var (
	ResponseTypeMap map[string]int = map[string]int{
		"ok": ResponseTypePass,
		"FAIL": ResponseTypeFail,
	}
	exclusions []string
	packageList []string
)

func getwd() string {
	directory, error := os.Getwd()
	if error != nil {
		fmt.Printf(errorCurrentDirectory, error)
		os.Exit(1)
	}
	return directory
}

func execute(command int, packageName string) bool {

	directory := getwd()

	if packageList == nil {
		var fileType string
		if command == CommandTest || command == CommandInstallTests {
			fileType = fileTypeTest
		} else {
			fileType = fileTypeGo
		}
		packageList = buildPackageList(directory, fileType)
		packageList = filterPackages(packageList, packageName, exclusions)
	}

	switch command {
	case CommandInstallTests:
		fmt.Print("\nInstalling Tests: ")
		// I cannot find example output for failed test install, so...
		// I'm leaving it like this for now.
		fmt.Print("\n\n")
		fmt.Print(installTests(packageList, directory))
	case CommandTest:
		fmt.Print("\nRunning tests: ")
		output := runShellCommand(directory, "go", makeArgs(packageList, "test")...)
		results := parseTestOutput(output)
		printSummary(results)
	case CommandInstall:
	case CommandRace:
		fmt.Print("\nRunning race tests: ")
		output := runShellCommand(directory, "go", makeArgs(packageList, "test", "-race")...)
		results := parseTestOutput(output)
		printSummary(results)
	case CommandVet:
		fmt.Print("\nVetting Packages: ")
		// This seems to just contain some basic output about
		// packages, so just print it out.
		fmt.Print("\n\n")
		fmt.Print(runShellCommand(directory, "go", makeArgs(packageList, "vet")...))
	case CommandCover:
		fmt.Print("\nRunning coverage tests: ")
		output := runShellCommand(directory, "go", makeArgs(packageList, "test", "-cover")...)
		results := parseTestOutput(output)
		printSummary(results)
		printCoverage(results)
	}

	return false

}

// installTests runs go test -i against the detected packages and
// returns the output.
func installTests(packageList []string, directory string) string {
	return runShellCommand(directory, "go", makeArgs(packageList, "test", "-i")...)
}

// responseType checks the response type of a line.
func responseType(line string) int {
	firstWord := ""
	fields := strings.Fields(line)
	if len(fields) > 0 {
		firstWord = strings.Fields(line)[0]
	}
	typeString := strings.TrimSpace(firstWord)
	if response, ok := ResponseTypeMap[typeString]; ok {
		return response
	}
	return -1
}

// parseTestOutput takes the output from running "go test" on a set of
// packages and parses it into information about the test results.
func parseTestOutput(commandOutput string) objx.Map {
	results := strings.Split(commandOutput, "\n")
	responseMap := objx.New(map[string]interface{}{
		"pass": []string{},
		"fail": []string{},
		"coverage": map[string]float32{},
	})
	var (
		currentMessage string
		coverage float32
		pkgName string
	)
	for _, result := range results {
		pkgName = ""
		coverage = -1.0
		switch responseType(result) {
		case ResponseTypePass:
			currentPasses := responseMap.Get("pass").StrSlice()
			responseMap.Set("pass", append(currentPasses, result))
			pkgName, coverage = parseNameAndCoverage(result)
		case ResponseTypeFail:
			currentMessage = fmt.Sprintf("%s\n%s", currentMessage, result)
			if result != "FAIL" {
				// Failure output has two lines - first, just "FAIL",
				// and second, the failing test summary.  When we hit
				// the summary, we want to store the message.
				currentFailures := responseMap.Get("fail").StrSlice()
				responseMap.Set("fail", append(currentFailures, currentMessage))
				currentMessage = ""
				pkgName, coverage = parseNameAndCoverage(result)
			}
		default:
			currentMessage = fmt.Sprintf("%s\n%s", currentMessage, result)
		}
		if pkgName != "" {
			coverageMap := responseMap["coverage"].(map[string]float32)
			coverageMap[pkgName] = coverage
		}
	}
	return responseMap
}

// parseCoverage takes a line and attempts to parse out the test
// package and coverage value (in percentage, out of 100), if it
// exists.  If no coverage value is found, a value of -1.0 will
// be returned.
func parseNameAndCoverage(line string) (string, float32) {
	pkgName := strings.Fields(line)[1]
	var coverage float32 = -1.0
	coverageLine := "coverage: "
	coverageStart := strings.Index(line, coverageLine)
	var coverageEnd int
	if coverageStart >= 0 {
		coverageStart += len(coverageLine)
		coverageEnd = coverageStart + strings.Index(line[coverageStart:], "%")
	}
	if coverageStart >= 0 && coverageEnd > 0 {
		coverage64, err := strconv.ParseFloat(line[coverageStart:coverageEnd], 32)
		if err == nil {
			coverage = float32(coverage64)
		}
	}
	return pkgName, coverage
}

// printSummary will print out a summary of the results, and if there
// were failures, it will print out output from the command(s) run.
func printSummary(results objx.Map) {
	failures := len(results.Get("fail").StrSlice())
	passes := len(results.Get("pass").StrSlice())
	tests := failures + passes
	if failures > 0 {
		fmt.Println("Passed Packages:")
		for _, passMessage := range results.Get("pass").StrSlice() {
			fmt.Print(passMessage)
			fmt.Println()
		}
		fmt.Println()
		fmt.Println("Failed Packages:")
		for _, failMessage := range results.Get("fail").StrSlice() {
			fmt.Print(failMessage)
			fmt.Println()
		}
		fmt.Println()
	}
	fmt.Printf("\n\n%d run. %d succeeded. %d failed. [%.0f%% success]\n\n",
		tests, passes, failures, float32(passes)/float32(tests)*100)
}

func printCoverage(results objx.Map) {
	fmt.Print("Coverage Summary: \n\n")
	for pkgName, coverage := range results["coverage"].(map[string]float32) {
		if coverage >= 0 {
			fmt.Printf("%s: %.1f%%\n", pkgName, coverage)
		} else {
			fmt.Printf("%s: N/A (tests failed or no tests found)\n", pkgName)
		}
	}
	fmt.Println()
}

// runShellCommand runs a shell command in a specified directory and returns
// a string containing all error output if the command fails
func runShellCommand(directory, command string, arguments ...string) string {
	shellCommand := exec.Command(command, arguments...)
	shellCommand.Dir = directory

	output, _ := shellCommand.CombinedOutput()
	return string(output)

}

func makeArgs(packages []string, commands ...string) []string {

	return append(commands, packages...)

}

func main() {

	var config = readConfig()
	exclusions = config[configKeyExclusions].([]string)

	commander.Go(func() {
		// The default command installs tests, then runs tests.
		commander.Map(commander.DefaultCommand, "", "",
			func(args objx.Map) {
				execute(CommandTest, "")
				return
				if execute(CommandInstallTests, "") {
					execute(CommandTest, "")
				}
			})

		commander.Map("test [packageName=(string)]", "Runs tests, or named test",
			"If no packageName argument is specified, runs all tests recursively. If a packageName argument is specified, runs just that test, unless the argument is \"all\", in which case it runs all tests, including those in the exclusion list.",
			func(args objx.Map) {
				//packageName := args.GetString("packageName")
				execute(CommandTest, "")
			})

		commander.Map("install [packageName=(string)]", "Installs tests, or named test",
			"If no packageName argument is specified, installs all tests recursively. If a packageName argument is specified, installs just that test, unless the argument is \"all\", in which case it installs all tests, including those in the exclusion list.",
			func(args objx.Map) {
				//packageName := args.GetString("packageName")
				execute(CommandInstallTests, "")
			})

		commander.Map("vet [packageName=(string)]", "Vets packageNames, or named packageName",
			"If no packageName argument is specified, vets all packageNames recursively. If a packageName argument is specified, vets just that packageName, unless the argument is \"all\", in which case it vets all packageNames, including those in the exclusion list.",
			func(args objx.Map) {
				//packageName := args.GetString("packageName")
				execute(CommandVet, "")
			})

		commander.Map("race [packageName=(string)]", "Runs race detector on tests, or named test",
			"If no packageName argument is specified, race tests all tests recursively. If a packageName argument is specified, vets just that test, unless the argument is \"all\", in which case it vets all tests, including those in the exclusion list.",
			func(args objx.Map) {
				//packageName := args.GetString("packageName")
				execute(CommandRace, "")
			})

		commander.Map("coverage [packageName=(string)]", "Runs the test coverage tool on tests, or a named test",
			"If no packageName argument is specified, coverage tests all tests recursively.  If a packageName argument is specified, checks coverage of just that package, unless the argument is \"all\", in which case it runs against all tests, including those in the exclusion list.",
			func(args objx.Map) {
				//packageName := args.GetString("packageName")
				execute(CommandCover, "")
			})

		commander.Map("exclude packageName=(string)", "Excludes the named directory from recursion",
			"An excluded directory will be skipped when walking the directory tree. Any subdirectories of the excluded directory will also be skipped.",
			func(args objx.Map) {
				packageName := args.Get("packageName").Str()
				exclude(packageName, config)
				fmt.Printf("\nExcluded \"%s\" from being examined during recursion.\n", packageName)
				config = readConfig()
				exclusions = config[configKeyExclusions].([]string)
				fmt.Printf("\n%s\n\n", formatExclusionsForPrint(exclusions))
			})

		commander.Map("include packageName=(string)", "Removes the named directory from the exclusion list", "",
			func(args objx.Map) {
				packageName := args.Get("packageName").Str()
				include(packageName, config)
				fmt.Printf("\nRemoved \"%s\" from the exclusion list.\n", packageName)
				fmt.Printf("\n%s\n\n", formatExclusionsForPrint(exclusions))
			})

		commander.Map("exclusions", "Prints the exclusion list", "",
			func(args objx.Map) {
				fmt.Printf("\n%s\n\n", formatExclusionsForPrint(exclusions))
			})
	})

}
