/*
- Copy *_test.before to temporary dir as *_test.go
- Read user answers to prompts from "prompt:x" comments in *_test.go
- Run test mocking user answers
- Compare resulting test code with *_test.after
- Compare test run output with *_test.output
*/
package assert_value_go

import (
	"github.com/smetana/assert_value_go/assertvalue"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"testing"
)

var tmpDir string

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

func TestSimple(t *testing.T) {
	copyPath("test/assertvalue_test.before", "assertvalue_test.go")
	prompts := getPrompts("assertvalue_test.go")
	cmd := exec.Command("go", "test", "-v", "assertvalue_test.go", "-args", "--", "-prompts="+prompts)
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(),
		"GOFLAGS=-mod=vendor",
	)
	out, err := cmd.Output()
	if err != nil {
		t.Log("assertvalue_test.go failed")
		t.Log(string(out))
		t.Fatal(err)
	}
	testCode, err := ioutil.ReadFile(tmpDir + "/assertvalue_test.go")
	if err != nil {
		t.Fatal(err)
	}
	assertvalue.File(t, string(testCode), "test/assertvalue_test.after")
	assertvalue.File(t, string(out), "test/assertvalue_test.output")
}
