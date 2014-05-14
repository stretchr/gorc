package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/howeyc/fsnotify"
	"github.com/stretchr/commander"
	"github.com/stretchr/objx"
)

// An enum defining the various operations that gorc can perform
const (
	commandInstallTests = iota
	commandTest
	commandInstall
	commandRace
	commandVet
	commandCover
	commandWatch
)

// An enum defining the response types expected from parsing the test output
const (
	responseTypePass = iota
	responseTypeFail
)

const (
	// When we receive a filesystem event, delay this amount before
	// running the requested command.  This is necessary because a
	// recursive delete will trigger testing before all of the files
	// are deleted, resulting in file not found errors.
	watchDelay string = "1s"
)

var (
	responseTypeMap = map[string]int{
		"ok":   responseTypePass,
		"FAIL": responseTypeFail,
	}
	exclusions  []string
	packageList []string
	watchedDirs map[string]bool
)

func getwd() string {
	directory, error := os.Getwd()
	if error != nil {
		fmt.Printf(errorCurrentDirectory, error)
		os.Exit(1)
	}
	return directory
}

// execute runs one of the available commands.  It accepts an objx.Map
// of options for the given command.
func execute(command int, options objx.Map) bool {

	directory := getwd()
	packageName := options.Get("packageName").Str()

	if packageList == nil {
		var fileType string
		if command == commandTest || command == commandInstallTests {
			fileType = fileTypeTest
		} else {
			fileType = fileTypeGo
		}
		packageList = buildPackageList(directory, fileType)
		packageList = filterPackages(packageList, packageName, exclusions)
	}

	switch command {
	case commandInstallTests:
		fmt.Print("\nInstalling Tests: ")
		// I cannot find example output for failed test install, so...
		// I'm leaving it like this for now.
		fmt.Print("\n\n")
		output := runShellCommand(directory, "go", makeArgs(packageList, "test", "-i")...)
		success := true
		for line := range output {
			if line != "" {
				fmt.Println(line)
				success = false
			}
		}
		return success
	case commandTest:
		fmt.Print("\nRunning tests: ")
		output := runShellCommand(directory, "go", makeArgs(packageList, "test")...)
		results := parseTestOutput(output)
		printSummary(results)
	case commandInstall:
	case commandRace:
		fmt.Print("\nRunning race tests: ")
		output := runShellCommand(directory, "go", makeArgs(packageList, "test", "-race")...)
		results := parseTestOutput(output)
		printSummary(results)
	case commandVet:
		fmt.Print("\nVetting Packages: ")
		// This seems to just contain some basic output about
		// packages, so just print it out.
		fmt.Print("\n\n")
		fmt.Print(runShellCommand(directory, "go", makeArgs(packageList, "vet")...))
	case commandCover:
		fmt.Print("\nRunning coverage tests: \n\n")
		output := runShellCommand(directory, "go", makeArgs(packageList, "test", "-cover")...)
		results := parseTestOutput(output)
		printCoverage(results)
	case commandWatch:
		// For now, just watch the test command.  We'll deal with more
		// stuff later.
		watchCommandStr := options.Get("command").Str()
		var command int
		switch watchCommandStr {
		case "":
			fallthrough
		case "test":
			command = commandTest
		case "vet":
			command = commandVet
		case "race":
			command = commandRace
		case "coverage":
			command = commandCover
		}
		watch(command, options)
	}

	return false

}

func watch(command int, options objx.Map) {
	done := make(chan bool)
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, syscall.SIGINT)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	go watchListener(command, watcher, done, interrupt, options)
	watchedDirs = buildSubdirMap(".")
	for dir := range watchedDirs {
		watcher.Watch(dir)
	}
	<-done
	fmt.Print("\nDone - exiting...\n")
	watcher.Close()
}

func runWatcherTests(command int, event *fsnotify.FileEvent, watcher *fsnotify.Watcher, options objx.Map, finished chan bool) {
	// The package list may change every time there's a file
	// change event in the directory, so rebuild it each time.
	packageList = buildPackageList(getwd(), fileTypeTest)
	execute(command, options)
	fmt.Print("\n----------------------------------\n")
	finished <- true
}

func handleCreatePath(watcher *fsnotify.Watcher, path string) {
	finfo, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error reading new file: %s", path)
		return
	}
	if finfo.IsDir() {
		for dir := range buildSubdirMap(path) {
			fmt.Println(watcher.Watch(dir))
			watchedDirs[dir] = true
		}
	}
}

func handleDeletePath(watcher *fsnotify.Watcher, path string) {
	if _, watched := watchedDirs[path]; watched {
		fmt.Printf("Unwatching %s\n", path)
		fmt.Println(watcher.RemoveWatch(path))
		delete(watchedDirs, path)
		for watchedDir := range watchedDirs {
			// Remove sub-paths
			pathWithTrailingSlash := path + string(os.PathSeparator)
			if strings.HasPrefix(watchedDir, pathWithTrailingSlash) {
				fmt.Println(watcher.RemoveWatch(watchedDir))
				delete(watchedDirs, watchedDir)
			}
		}
	}
}

func handleEvent(watcher *fsnotify.Watcher, event *fsnotify.FileEvent) {
	path := event.Name
	if strings.HasPrefix(path, "./") {
		path = path[2:]
	}
	fmt.Printf("Received event for path %s: ", path)
	switch {
	case event.IsRename():
		fmt.Print("Renamed\n")
		if _, isWatched := watchedDirs[path]; isWatched {
			handleDeletePath(watcher, path)
		} else {
			handleCreatePath(watcher, path)
		}
	case event.IsDelete():
		fmt.Print("Deleted\n")
		handleDeletePath(watcher, path)
	case event.IsCreate():
		fmt.Printf("Created\n")
		handleCreatePath(watcher, path)
	}
}

func watchListener(command int, watcher *fsnotify.Watcher, doneChan chan bool, interruptChan chan os.Signal, options objx.Map) {
	fmt.Printf("\nStarting FS Watcher for current directory and sub-"+
		"directories, and running %s tests whenever files are changed...", options.Get("command").Str())
	fmt.Print("\n\n")
	fmt.Print("\n----------------------------------\n")
	testFinished := make(chan bool)
	testing := false
	delayDuration, err := time.ParseDuration(watchDelay)
	if err != nil {
		panic(err)
	}
	var (
		delayTimer <-chan time.Time
		event      *fsnotify.FileEvent
	)
	for {
		select {
		case event = <-watcher.Event:
			handleEvent(watcher, event)
			delayTimer = time.After(delayDuration)
		case <-delayTimer:
			if !testing {
				testing = true
				go runWatcherTests(command, event, watcher, options, testFinished)
			}
		case <-testFinished:
			testing = false
		case <-interruptChan:
			doneChan <- true
			break
		case err := <-watcher.Error:
			fmt.Printf("Error: %s\n", err)
			panic(err)
		}
	}
}

// responseType checks the response type of a line.
func responseType(line string) int {
	firstWord := ""
	fields := strings.Fields(line)
	if len(fields) > 0 {
		firstWord = strings.Fields(line)[0]
	}
	typeString := strings.TrimSpace(firstWord)
	if response, ok := responseTypeMap[typeString]; ok {
		return response
	}
	return -1
}

// parseTestOutput takes the output from running "go test" on a set of
// packages and parses it into information about the test results.
func parseTestOutput(commandOutput chan string) objx.Map {
	responseMap := objx.New(map[string]interface{}{
		"pass":     []string{},
		"fail":     []string{},
		"coverage": map[string]float32{},
	})
	var (
		currentMessage string
		coverage       float32
		pkgName        string
	)
	for result := range commandOutput {
		pkgName = ""
		coverage = -1.0
		switch responseType(result) {
		case responseTypePass:
			currentPasses := responseMap.Get("pass").StrSlice()
			responseMap.Set("pass", append(currentPasses, result))
			pkgName, coverage = parseNameAndCoverage(result)
			printProgress(responseMap)
		case responseTypeFail:
			currentMessage = fmt.Sprintf("%s\n%s", currentMessage, result)
			if result != "FAIL" {
				// Failure output has two lines - first, just "FAIL",
				// and second, the failing test summary.  When we hit
				// the summary, we want to store the message.
				currentFailures := responseMap.Get("fail").StrSlice()
				responseMap.Set("fail", append(currentFailures, currentMessage))
				currentMessage = ""
				pkgName, coverage = parseNameAndCoverage(result)
				printProgress(responseMap)
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
	fields := strings.Fields(line)
	if len(fields) <= 1 {
		return "", -1.0
	}
	pkgName := fields[1]
	coverage := -1.0
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
			coverage = coverage64
		}
	}
	return pkgName, float32(coverage)
}

var lastPrintLen = 0

// printProgress will print out the progress of these tests as they run
func printProgress(results objx.Map) {
	failures := len(results.Get("fail").StrSlice())
	passes := len(results.Get("pass").StrSlice())
	tests := failures + passes

	if lastPrintLen == 0 {
		printString := fmt.Sprintf("[%d of %d]", tests, len(packageList))
		lastPrintLen = len(printString)
		fmt.Print(printString)
	} else {
		printString := fmt.Sprintf("%s[%d of %d]", strings.Repeat("\b", lastPrintLen), tests, len(packageList))
		lastPrintLen = len(printString) - lastPrintLen
		fmt.Print(printString)
	}

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
func runShellCommand(directory, command string, arguments ...string) chan string {
	shellCommand := exec.Command(command, arguments...)
	shellCommand.Dir = directory
	stdoutPipe, err := shellCommand.StdoutPipe()
	if err != nil {
		panic(fmt.Sprintf("Cannot establish pipe to command: %s", err))
	}
	stderrPipe, err := shellCommand.StderrPipe()
	if err != nil {
		panic(fmt.Sprintf("Cannot establish pipe to command: %s", err))
	}

	output := make(chan string)

	go func() {
		shellCommand.Start()
		stdoutReader := bufio.NewReader(stdoutPipe)
		stderrReader := bufio.NewReader(stderrPipe)
		for {
			stdoutLine, stdoutErr := stdoutReader.ReadString('\n')
			stderrLine, stderrErr := stderrReader.ReadString('\n')

			if stdoutErr != nil && stderrErr != nil {
				if len(stdoutLine) > 0 {
					output <- stdoutLine
				}
				if len(stderrLine) > 0 {
					output <- stderrLine
				}
				close(output)
				return
			}
			if len(stdoutLine) > 0 {
				output <- stdoutLine
			}
			if len(stderrLine) > 0 {
				output <- stderrLine
			}
		}
	}()
	return output
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
				if execute(commandInstallTests, args) {
					execute(commandTest, args)
				}
			})

		commander.Map("test [packageName=(string)]", "Runs tests, or named test",
			"If no packageName argument is specified, runs all tests recursively. If a packageName argument is specified, runs just that test, unless the argument is \"all\", in which case it runs all tests, including those in the exclusion list.",
			func(args objx.Map) {
				//packageName := args.GetString("packageName")
				execute(commandTest, args)
			})

		commander.Map("install [packageName=(string)]", "Installs tests, or named test",
			"If no packageName argument is specified, installs all tests recursively. If a packageName argument is specified, installs just that test, unless the argument is \"all\", in which case it installs all tests, including those in the exclusion list.",
			func(args objx.Map) {
				//packageName := args.GetString("packageName")
				execute(commandInstallTests, args)
			})

		commander.Map("vet [packageName=(string)]", "Vets packageNames, or named packageName",
			"If no packageName argument is specified, vets all packageNames recursively. If a packageName argument is specified, vets just that packageName, unless the argument is \"all\", in which case it vets all packageNames, including those in the exclusion list.",
			func(args objx.Map) {
				//packageName := args.GetString("packageName")
				execute(commandVet, args)
			})

		commander.Map("race [packageName=(string)]", "Runs race detector on tests, or named test",
			"If no packageName argument is specified, race tests all tests recursively. If a packageName argument is specified, vets just that test, unless the argument is \"all\", in which case it vets all tests, including those in the exclusion list.",
			func(args objx.Map) {
				//packageName := args.GetString("packageName")
				execute(commandRace, args)
			})

		commander.Map("coverage [packageName=(string)]", "Runs the test coverage tool on tests, or a named test",
			"If no packageName argument is specified, coverage tests all tests recursively.  If a packageName argument is specified, checks coverage of just that package, unless the argument is \"all\", in which case it runs against all tests, including those in the exclusion list.",
			func(args objx.Map) {
				//packageName := args.GetString("packageName")
				execute(commandCover, args)
			})

		commander.Map("watch [command=(test)] [packageName=(string)]", "Watch for file changes and run gorc test every time a file changes",
			"If no packageName argument is specified, watch tests recursively.  If a packageName argument is specified, watches just that package, unless the argument is \"all\", in which case it watches all packages, including those in the exclusion list.",
			func(args objx.Map) {
				// packageName := args.GetString("packageName")
				execute(commandWatch, args)
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
