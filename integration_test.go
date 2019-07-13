/*
- Copy *_test.before to temporary dir as *_test.go
- Read user answers to prompts from "prompt:x" comments in *_test.go
- Run test mocking user answers
- Compare resulting test code with *_test.after
- Compare test run output with *_test.output
*/
package assert_value_go

import (
	"bytes"
	"github.com/smetana/assert_value_go/assertvalue"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"syscall"
	"testing"
)

var (
	tmpDir   string
	canonRe1 *regexp.Regexp
	canonRe2 *regexp.Regexp
)

// ----------------- Tests -------------------

func TestPass(t *testing.T) {
	runTestFile(t, "assertvalue_test", true)
}

func TestFail(t *testing.T) {
	runTestFile(t, "fail_test", false)
}

// ----------------- Helpers -----------------

func init() {
	canonRe1 = regexp.MustCompile(`((ok|FAIL)\s+command-line-arguments\s*)(.*)`)
	canonRe2 = regexp.MustCompile(`((PASS|FAIL):\s+Test.*\s+)\(.*?\)`)
}

func TestMain(m *testing.M) {
	var err error
	tmpDir, err = ioutil.TempDir("", "assertvalue_test")
	if err != nil {
		log.Fatal(err)
	}
	pathsToCopy := []string{
		"go.mod",
		"go.sum",
		"assertvalue",
		"vendor",
	}
	for _, path := range pathsToCopy {
		copyPath(path, "")
	}

	code := m.Run()
	os.RemoveAll(tmpDir)
	os.Exit(code)
}

func runTestFile(t *testing.T, testName string, shouldPass bool) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	beforeFilename := "test/" + testName + ".before"
	afterFilename := "test/" + testName + ".after"
	outputFilename := "test/" + testName + ".output"
	testFilename := testName + ".go"

	copyPath(beforeFilename, testFilename)
	prompts := getPrompts(testFilename)
	cmd := exec.Command("go", "test", "-v", testFilename,
		"-args", "--", "-prompts="+prompts,
	)
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(),
		"GOFLAGS=-mod=vendor",
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		if shouldPass {
			t.Log(testFilename + " failed")
			t.Log(stdout.String())
			t.Log(stderr.String())
			t.Fatal(err)
		} else {
			if exitError, ok := err.(*exec.ExitError); ok {
				if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
					if status.ExitStatus() != 1 {
						t.Log(testFilename + " failed")
						t.Log(stdout.String())
						t.Log(stderr.String())
						t.Fatal(err)
					}
				}
			}
		}
	}
	testCode, err := ioutil.ReadFile(tmpDir + "/" + testFilename)
	if err != nil {
		t.Fatal(err)
	}
	out := canonicalizeOutput(stdout.String())
	assertvalue.File(t, string(testCode), afterFilename)
	assertvalue.File(t, out, outputFilename)
}

func copyPath(in, out string) {
	cmd := exec.Command("cp", "-r", in, tmpDir+"/"+out)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func getPrompts(in string) string {
	testCode, err := ioutil.ReadFile(tmpDir + "/" + in)
	if err != nil {
		log.Fatal(err)
	}
	re := regexp.MustCompile(`prompt[\s:=]*([ynYN]*)`)
	matches := re.FindAllStringSubmatch(string(testCode), -1)
	prompts := ""
	for _, match := range matches {
		prompts = prompts + match[1]
	}
	return prompts
}

func canonicalizeOutput(s string) string {
	s = canonRe1.ReplaceAllString(s, "${1}0000s")
	s = canonRe2.ReplaceAllString(s, "${1}(0000s)")
	return s
}
