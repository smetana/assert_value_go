package assertvalue

import (
	"bufio"
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"github.com/pmezard/go-difflib/difflib"
	"log"
	"os"
	"runtime"
	"strings"
	"testing"
	"unicode"
)

var FileData = "Hey!"

const maxInt = int(^uint(0) >> 1)

const banner = `
For now assertvalue.Equal supports only two forms:

  asertvalue.Equal(t, actual)

  assertvalue.Equal(t, actual, D(` + "`" + `
    ..expected value..
  ` + "`" + `))

`

func Equal(t *testing.T, args ...string) {
	var actual, expected string

	if len(args) == 1 {
		actual = args[0]
		expected = ""
	} else if len(args) == 2 {
		actual = args[0]
		expected = args[1]
	} else {
		t.Fatal("Invalid function call\n" + banner)
	}

	diff := difflib.UnifiedDiff{
		A:       difflib.SplitLines(expected),
		B:       difflib.SplitLines(actual),
		Context: 3,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)
	fmt.Println(text)
	fmt.Println(t.Name())
	_, file, line, _ := runtime.Caller(1)
	fmt.Println(line)
	lines := readLines(file)
	fmt.Println("---")
	fmt.Println(lines[line-1])
	fmt.Println("---")
	fmt.Println(formatNewExpected(actual))
	t.Fatal("not implemented")
}

func readLines(filename string) []string {
	var lines []string
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	err = s.Err()
	if err != nil {
		log.Fatal(err)
	}
	return lines
}

func getIndentation(s string) int {
	lines := strings.Split(s, "\n")
	indent := maxInt
	for _, line := range lines {
		lineIndent := 0
		for _, r := range []rune(line) {
			if unicode.IsSpace(r) {
				lineIndent += 1
			} else {
				break
			}
		}

		if lineIndent < indent {
			indent = lineIndent
		}
	}
	return indent
}

func formatNewExpected(s string) string {
	return "D(`\n" + heredoc.Docf(s) + "`)"
}
