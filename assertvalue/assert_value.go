package assertvalue

import (
	"bufio"
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"github.com/pmezard/go-difflib/difflib"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

const maxInt = int(^uint(0) >> 1)

var (
	defaultAnswer      string
	isInteractive      = true
	acceptNewValues    = false
	reStringNoExpected = regexp.MustCompile(`^(\s*)(assertvalue\.String\(.*)(\))`)
)

func init() {
	for _, arg := range os.Args {
		switch arg {
		case "nointeractive":
			isInteractive = false
		case "accept":
			acceptNewValues = true
		}
	}
}

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

func String(t *testing.T, args ...string) {
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

			For now assertvalue.String supports only two forms:

			asertvalue.Equal(t, actual)

		    assertvalue.Equal(t, actual, D(` + "`" + `
		      ..expected value..
		    ` + "`" + `))

		`))
	}

	// Since we use heredocs to store values and heredoc always ends with
	// new line character we should add do something with values
	// which don't end wuth new lines. For now indicate missing new line
	// with <NOEOL>\n
	if actual[len(actual)-1:] != "\n" {
		actual = actual + "<NOEOL>\n"
	}

	if actual != expected {
		diffStruct := difflib.UnifiedDiff{
			A:       difflib.SplitLines(expected),
			B:       difflib.SplitLines(actual),
			Context: 3,
		}
		diff, _ := difflib.GetUnifiedDiffString(diffStruct)
		if !isNewValueAccepted(diff) {
			t.FailNow()
		} else {
			_, filename, lineno, _ := runtime.Caller(1)
			code := readTestCode(filename)
			line := code[lineno-1]
			parsed := reStringNoExpected.FindAllStringSubmatch(line, -1)
			indent := parsed[0][1]
			prefix := parsed[0][2]
			suffix := parsed[0][3]
			expected = formatNewExpected(actual, indent)
			code[lineno-1] = indent + prefix + ", " + expected + suffix
			writeTestCode(filename, code)
		}
	}
}

func isNewValueAccepted(diff string) bool {
	fmt.Println(diff)
	var answer string
	if isInteractive {
		if defaultAnswer != "" {
			answer = defaultAnswer
		} else {
			fmt.Print("Accept new value? [y,n,Y,N] ")
			// testing framework changes os.Stdin
			// We need real interaction with user
			tty, err := os.Open("/dev/tty")
			if err != nil {
				log.Fatalf("can't open /dev/tty: %s", err)
			}
			s := bufio.NewScanner(tty)
			s.Scan()
			answer = s.Text()
			if answer == "Y" || answer == "N" {
				defaultAnswer = answer
			}
		}
		return answer == "y" || answer == "Y"
	} else if acceptNewValues {
		return true
	} else {
		return false
	}
}

func readTestCode(filename string) []string {
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

func writeTestCode(filename string, code []string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, line := range code {
		fmt.Fprintln(w, line)
	}
	err = w.Flush()
	if err != nil {
		log.Fatal(err)
	}
}

func formatNewExpected(s, indent string) string {
	return "D(`\n" + formatExpectedContent(s, indent) + indent + "`)"
}

func formatExpectedContent(s, indent string) string {
	expected := heredoc.Docf(s)
	lines := strings.Split(expected, "\n")
	for i, line := range lines {
		if i < len(lines)-1 {
			lines[i] = indent + "\t" + line
		}
	}
	return strings.Join(lines, "\n")
}
