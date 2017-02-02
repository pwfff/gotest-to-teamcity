# gotest-to-teamcity
Translate `go test -v` output into TeamCity service messages.

[TeamCity](https://www.jetbrains.com/teamcity/) works by monitoring the
output of a test script, looking for specially formatted ["service messages"](https://confluence.jetbrains.com/display/TCD10/Build+Script+Interaction+with+TeamCity).
These service messages look like this:
```sh
##teamcity[testStarted name='testName'
##teamcity[testFinished name='testName' duration='<test_duration_in_milliseconds>']
##teamcity[testFailed name='MyTest.test1' message='failure message' details='message and stack trace']
```
This is how TeamCity records passes, fails, and skips.

Unfortunately, JetBrains does not provide a script for generating service
messages from `go test -v` output. So that's what this does.

**IMPORTANT: You must pass the `-v` flag to `go test`, or `gotest-to-teamcity` will not work.**

### Usage
```
go get -u github.com/apcera/gotest-to-teamcity

go test -v ./... | gotest-to-teamcity  # must pass the -v flag!
```
