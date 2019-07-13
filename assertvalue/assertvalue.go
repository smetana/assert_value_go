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
	// See init() for comments
	recurringAnswer string
	isInteractive   bool
	acceptNewValues bool
	reCodeNoExp     *regexp.Regexp
	reCodeExpBegin  *regexp.Regexp
	reCodeExpEnd    *regexp.Regexp
	fileChanges     map[string]map[int]int
	prompts         []string
)

func init() {
	isInteractive = true
	acceptNewValues = false
	for _, arg := range os.Args {
		switch arg {
		case "-nointeractive":
			isInteractive = false
		case "-accept":
			acceptNewValues = true
		}
	}
	// Parse "-promts nnnyyy" or "-prompts=nnnyyy" argument
	rePrompts := regexp.MustCompile(`-prompts(\s+|=)(.*)(\s|$)`)
	parsed := rePrompts.FindAllStringSubmatch(strings.Join(os.Args, " "), -1)
	if len(parsed) > 0 {
		prompts = strings.Split(parsed[0][2], "")
	}
	// Match assertValie.String call without expected
	reCodeNoExp = regexp.MustCompile(`^(\s*)(assertvalue\.String\(.*)(\))`)
	// Match first line of assertValie.String call with expected following
	reCodeExpBegin = regexp.MustCompile("^(\\s*)(assertvalue\\.String\\(.*,\\s*`)$")
	// Match closing `) after expected value in assertvalue.String call
	// Compliment to reCodeExpBegin
	reCodeExpEnd = regexp.MustCompile("^\\s*`\\s*\\)")
	// Keep tracking of changes in test code
	// Changing expected may change the number of lines in test code
	// and runtime.Caller returns initial file line numbers
	// We keep line number changes here in the form of offsets
	// fileChanges[filename][line] => offset
	fileChanges = make(map[string]map[int]int)
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
		expected = heredoc.Doc(args[1])
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
	if actual == "" || actual[len(actual)-1:] != "\n" {
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
			_, filename, lineNum, _ := runtime.Caller(1)
			lineNumOrig := lineNum
			lineNum = currentLineNumber(filename, lineNum)
			code := readTestCode(filename)
			offset := 0
			if len(args) == 1 {
				code, offset = createExpected(code, lineNum, actual)
			} else {
				code, offset = updateExpected(code, lineNum, actual, t)
			}
			if offset != 0 {
				updateLineNumbers(filename, lineNumOrig, offset)
			}
			writeTestCode(filename, code)
		}
	}
}

func isNewValueAccepted(diff string) bool {
	fmt.Println(diff)
	var answer string
	if isInteractive && testing.Verbose() {
		if recurringAnswer != "" {
			answer = recurringAnswer
		} else {
			fmt.Print("Accept new value? [y,n,Y,N] ")
			if len(prompts) > 0 {
				answer, prompts = prompts[0], prompts[1:]
				fmt.Println(answer)
			} else {
				// testing framework changes os.Stdin
				// We need real interaction with user
				tty, err := os.Open("/dev/tty")
				if err != nil {
					log.Fatalf("can't open /dev/tty: %s", err)
				}
				s := bufio.NewScanner(tty)
				s.Scan()
				answer = s.Text()
			}
			if answer == "Y" || answer == "N" {
				recurringAnswer = answer
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

func createExpected(code []string, lineNum int, actual string) ([]string, int) {
	line := code[lineNum-1]
	parsed := reCodeNoExp.FindAllStringSubmatch(line, -1)
	indent := parsed[0][1]
	prefix := parsed[0][2]
	suffix := parsed[0][3]
	expected := "`\n" +
		formatExpectedContent(actual, indent) +
		"\n" +
		indent +
		"`"
	code[lineNum-1] = indent + prefix + ", " + expected + suffix
	offset := strings.Count(expected, "\n")
	return code, offset
}

func updateExpected(code []string, lineNum int, actual string, t *testing.T) ([]string, int) {
	line := code[lineNum-1]
	parsed := reCodeExpBegin.FindAllStringSubmatch(line, -1)
	if len(parsed) != 1 {
		t.Fatal(`Unable to parse expected from string` + "\n" + line)
	}
	indent := parsed[0][1]
	// expected line number start/end
	expStart := lineNum
	expEnd := 0
	for expEnd == 0 && lineNum < len(code) {
		if reCodeExpEnd.Match([]byte(code[lineNum])) {
			expEnd = lineNum
		} else {
			lineNum = lineNum + 1
		}
	}
	expected := formatExpectedContent(actual, indent)
	newCode := make([]string, expStart)
	copy(newCode, code[:expStart])
	newCode = append(newCode, expected)
	newCode = append(newCode, code[expEnd:]...)
	offset := strings.Count(expected, "\n") - (expEnd - expStart - 1)
	return newCode, offset
}

func formatExpectedContent(s, indent string) string {
	expected := heredoc.Docf(s)
	lines := strings.Split(expected, "\n")
	// cut empty last line
	lines = lines[:len(lines)-1]
	for i, line := range lines {
		lines[i] = indent + "\t" + line
	}
	return strings.Join(lines, "\n")
}

func updateLineNumbers(filename string, lineNum, offset int) {
	if fileChanges[filename] == nil {
		fileChanges[filename] = make(map[int]int)
	}
	fileChanges[filename][lineNum] = offset
}

func currentLineNumber(filename string, lineNum int) int {
	cumulativeOffset := 0
	if fileChanges[filename] != nil {
		for num, offset := range fileChanges[filename] {
			if lineNum > num {
				cumulativeOffset = cumulativeOffset + offset
			}
		}
	}
	return lineNum + cumulativeOffset
}
