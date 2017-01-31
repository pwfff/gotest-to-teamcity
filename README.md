# gotest-to-teamcity
Translate `go test -v` output into TeamCity service messages.

### Usage
```sh
go test -v ./... | gotest-to-teamcity
```

**IMPORTANT: `gotest-to-teamcity` expectes the verbose output from `go test`.
You must pass the `-v` flag to `go test`, or `gotest-to-teamcity` will fail.**
