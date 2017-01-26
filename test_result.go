// Copyright 2016 Apcera Inc. All rights reserved.

// FIXME: This should be a "result" subpackage, but I don't have permission to
// make a separate gotest-to-teamcity repo, so into main it goes.
package main

import (
	"fmt"
	"strings"
	"time"
)

// TeamCity service messages have a special syntax for escaping characters.
// https://confluence.jetbrains.com/display/TCD10/Build+Script+Interaction+with+TeamCity#BuildScriptInteractionwithTeamCity-Escapedvalues
var replacer = strings.NewReplacer(
	"|", "||",
	"'", "|'",
	"[", "|[",
	"]", "|]",
	"\n", "|n",
	"\r", "|r",
	"\u0085", "|x",
	"\u2028", "|l",
	"\u2029", "|p",
)

type Result struct {
	TestName     string
	PassFailSkip string // "PASS", "FAIL", or "SKIP"
	Duration     time.Duration
	FailLogs     []string // the lines after "--- FAIL". empty if test passed.
}

// New returns a Result. If the given duration is invalid, the test's duration
// will be 0.
func New(name, passFailSkip, duration string) Result {
	// If an error occurs, a duration of 0 will be printed. This is ok.
	d, _ := time.ParseDuration(duration)

	return Result{
		TestName:     name,
		PassFailSkip: passFailSkip,
		Duration:     d,
	}
}

// String returns the TeamCity service message representation of the test
// result, bookended by testStarted and testFinished service messages.
// An example of the output is:
//   ##teamcity[testStarted name='TestFoo']
//   ##teamcity[testFailed name='TestFoo']
//   ##teamcity[testFinished name='TestFoo' duration='123' details='bad stuff happened']
func (result Result) String() string {
	// A subtest could contain a special character in its name.
	result.TestName = replacer.Replace(result.TestName)

	s := fmt.Sprintf("##teamcity[testStarted name='%s']\n", result.TestName)

	switch result.PassFailSkip {
	case "PASS":
		// print nothing
	case "FAIL":
		s += result.testFailedMsg()
	case "SKIP":
		s += fmt.Sprintf("##teamcity[testIgnored name='%s']\n", result.TestName)
	default:
		msg := fmt.Sprintf(`PassFailSkip can only be "PASS", "FAIL", or "SKIP", not %q`, result.PassFailSkip)
		panic(msg)
	}

	// TC requires the duration be given in milliseconds.
	ms := int(result.Duration.Seconds() * 1000)
	s += fmt.Sprintf("##teamcity[testFinished name='%s' duration='%d']", result.TestName, ms)
	return s
}

// Empty returns true if the test result hasn't been populated with any data
// yet.
func (result *Result) Empty() bool {
	return result.TestName == ""
}

// testFailedMsg generates a TeamCity testFailed service message from the given
// test result.
func (result Result) testFailedMsg() string {
	// If the test failed, append "details='<failure-msg>'".
	details := ""
	if len(result.FailLogs) > 0 {
		failMsg := replacer.Replace(strings.Join(result.FailLogs, ""))
		details = fmt.Sprintf(" details='%s'", failMsg)
	}

	return fmt.Sprintf("##teamcity[testFailed name='%s'%s]", result.TestName, details)
}
