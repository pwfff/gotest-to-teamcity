// Copyright 2016 Apcera Inc. All rights reserved.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/apcera/gotest-to-teamcity/test"
)

var (
	//                                             RESULT        TEST NAME      DURATION
	testResultRx = regexp.MustCompile(`\s*--- (PASS|FAIL|SKIP): ([-\w/:\.]+) \((\d+\.\d+s)\)\s*$`)

	//                                      RESULT        PACKAGE
	pkgSummaryRx = regexp.MustCompile(`^\s*(?:ok|FAIL)\s+([-\w/.]+)`)

	testCmd = flag.String("run", "", "'go test' command to execute.")
)

// Sample `go test -v` output:
//     --- PASS: TestFoo (0.01s)
//       foo_test.go:91: this is a log message
//       foo_test.go:95: this is another log message
//     --- PASS: TestBar (0.01s)
//     --- PASS: TestBaz (0.12s)
//     ok	github.com/apcera/foo	0.014s

// gotest_to_teamcity expects the output of `go test -v` to be piped to it. It
// reads the input from stdin line by line. When it encounters a line that
// starts with "--- (PASS|FAIL|SKIP)", it prints a TeamCity service message,
// which looks like this:
//     ##teamcity[testStarted name='TestFoo']
// TeamCity monitors stdout and looks for these messages to record passes,
// fails, test duration, etc.

func main() {
	flag.Parse()

	if *testCmd != "" {
		run(*testCmd, os.Stdout)
	} else {
		process(bufio.NewReader(os.Stdin), os.Stdout)
	}
}

func run(testCmd string, output io.Writer) {
	args := strings.Split(testCmd, " ")
	cmd := exec.Command(args[0], args[1:]...)

	pipe, err := cmd.StdoutPipe()
	input := bufio.NewReader(pipe)
	if err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	process(input, output)

	switch exit := cmd.Wait().(type) {
	case *exec.ExitError:
		os.Exit(exit.ExitCode())
	default:
		panic(exit.Error())
	}
}

func process(input *bufio.Reader, output io.Writer) {
	var (
		currResult   test.Result
		pkgResults   []test.Result // all the test results for a single pkg
		startFailMsg bool          // record subsequent lines as fail msg from previous test
	)

	for {
		line, err := input.ReadString('\n')
		fmt.Fprint(output, line)
		if err != nil {
			// We will not get an EOF until `go test -v` finishes writing and
			// closes the other end of the pipe.
			break
		}

		if skipLine(line) {
			// The line is just noise that's safe to ignore.
			continue
		}

		// Have we completed all the tests for a single package?
		if match := pkgSummaryRx.FindStringSubmatch(line); len(match) > 0 {
			startFailMsg = false
			pkgResults = append(pkgResults, currResult)
			currResult = test.Result{}

			pkgName := match[1]
			printPkgResults(output, pkgName, pkgResults)
			pkgResults = pkgResults[:0]
			continue
		}

		// Is this a "--- (PASS|FAIL|SKIP)" line?
		if match := testResultRx.FindStringSubmatch(line); len(match) > 0 {
			if line := strings.TrimSpace(line); strings.HasPrefix(line, "--- FAIL") {
				startFailMsg = true
			} else {
				startFailMsg = false
				if !currResult.Empty() {
					// The very first test result will be empty at this point.
					pkgResults = append(pkgResults, currResult)
				}
			}

			name, passFailSkip, duration := match[2], match[1], match[3]
			currResult = test.New(name, passFailSkip, duration)
			continue
		}

		if startFailMsg {
			currResult.FailLogs = append(currResult.FailLogs, line)
		}
	}
}

// skipLine returns true if the line starts with "=== RUN" or "?", or is just
// "PASS" or "FAIL" with no other information.
func skipLine(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "=== RUN") ||
		strings.HasPrefix(line, "?") ||
		line == "PASS" ||
		line == "FAIL"
}

// printPkgResults prints all the TC service messages for a package's tests.
func printPkgResults(output io.Writer, pkgName string, pkgResults []test.Result) {
	fmt.Fprintf(output, "##teamcity[testSuiteStarted name='%s']\n", pkgName)
	for _, result := range pkgResults {
		fmt.Fprintln(output, result)
	}
	fmt.Fprintf(output, "##teamcity[testSuiteFinished name='%s']\n", pkgName)
}
