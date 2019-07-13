# assertvalue

`assert` on steroids that writes and updates tests for you.
Golang implementation of https://github.com/assert-value

## Usage

You can start with no expected value

```go
// example_test.go
package mypackage

import (
	"github.com/smetana/assert_value_go/assertvalue"
	"testing"
)

func TestExample(t *testing.T) {
	assertvalue.String(t, "Hello World!\n")
}
```
Run test in verbose mode
```
go test -v example_test.go
```
It will ask you about diff

```
=== RUN   TestExample
@@ -1 +1,2 @@
+Hello World!


Accept new value? [y,n,Y,N] y
--- PASS: TestExample (2.30s)
PASS
ok  	command-line-arguments	2.306s
```

If you answer "n" the test will fail, but if you answer "y" then expected value
will be created (or updated) from actual value and test code will be changed

```go
// example_test.go
package assert_value_go

import (
	"github.com/smetana/assert_value_go/assertvalue"
	"testing"
)

func TestExample(t *testing.T) {
	assertvalue.String(t, "Hello World!\n", `
		Hello World!
	`)
}
```

## Features and Known Issues

### String literals as heredocs

Expected values are stored in string literals which are treated as heredocs for
better diff readability. This means they ignore common indentation and always
end with newlines.  When expected value does not end with a newline
`assertvalue` will append a special ```<NOEOL>``` string to indicate that last
newline should be ignored.
```go
assertvalue.String(t, "Hello World!", `
	Hello World!<NOEOL>
`)
```


### Running tests interactively and non-interactively

`assertvalue` interacts with user only in verbose mode, when `go test` is
executed with -v flag
```
# run test interactively in verbose mode
go test -v example_test.go
```
In normal mode `assertvalue` works as simple assert function and will fail when
expected value not equal actual value
```
# run test non-interactively in normal mode
go test example_test.go
```
You can also run test noninteractively in verbose mode using `-nointerative`
command line argument which may be useful for CI testing

```
# run test non-interactively in verbose mode
go test -v example_test -args -- -nointeractive
```

## API

For now this package is primitive and supports only `string` expected and
actual values

### assertvalue.String

Supports two forms
```go
assertvalue.String(t *testing.T, actual string)
assertvalue.String(t *testing.T, actual, expected string)
```
### assertvalue.File

If expected values are big to store them in test code you
can store them in files (hello .golden)

```go
assertvalue.File(t *testing.T, actual, filename string)
```
