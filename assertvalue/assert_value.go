package assertvalue

import (
	"bufio"
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"github.com/pmezard/go-difflib/difflib"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"testing"
	"unicode"
)

const maxInt = int(^uint(0) >> 1)

func File(t *testing.T, actual, filename string) {
	var expected string
	if _, err := os.Stat(filename); err == nil {
		// File exists. Use content as expected value
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}
		expected = string(buf)
	} else if os.IsNotExist(err) {
		// File does not exist. Will create file
		expected = ""
	} else {
		// Something happened
		log.Fatal(err)
	}
	if actual != expected {
		diffStruct := difflib.UnifiedDiff{
			A:        difflib.SplitLines(expected),
			B:        difflib.SplitLines(actual),
			FromFile: "file: " + filename,
			ToFile:   "actual",
			Context:  3,
		}
		diff, _ := difflib.GetUnifiedDiffString(diffStruct)
		if !isNewValueAccepted(diff) {
			t.FailNow()
		} else {
			err := ioutil.WriteFile(filename, []byte(actual), 0644)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func Equal(t *testing.T, args ...string) {
	var actual, expected string

	if len(args) == 1 {
		actual = args[0]
		expected = ""
	} else if len(args) == 2 {
		actual = args[0]
		expected = args[1]
	} else {
		t.Fatal(heredoc.Doc(`
			Invalid function call

			For now assertvalue.Equal supports only two forms:

			asertvalue.Equal(t, actual)

		    assertvalue.Equal(t, actual, D(` + "`" + `
		      ..expected value..
		    ` + "`" + `))

		`))
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

func isNewValueAccepted(diff string) bool {
	fmt.Println(diff)
	fmt.Print("Accept new value? [y,n] ")
	var answer string
	// testing framework changes os.Stdin
	// We need real interaction with user
	tty, err := os.Open("/dev/tty")
	if err != nil {
		log.Fatalf("can't open /dev/tty: %s", err)
	}
	s := bufio.NewScanner(tty)
	s.Scan()
	answer = s.Text()
	return answer == "y" || answer == "Y"
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
