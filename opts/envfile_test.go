package opts

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func tmpFileWithContent(content string, t *testing.T) string {
	tmpFile, err := ioutil.TempFile("", "envfile-test")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	tmpFile.WriteString(content)
	return tmpFile.Name()
}

// Test ParseEnvFile for a file with a few well formatted lines
func TestParseEnvFileGoodFile(t *testing.T) {
	content := `foo=bar
    baz=quux
# comment

_foobar=foobaz
with.dots=working
and_underscore=working too
`
	// Adding a newline + a line with pure whitespace.
	// This is being done like this instead of the block above
	// because it's common for editors to trim trailing whitespace
	// from lines, which becomes annoying since that's the
	// exact thing we need to test.
	content += "\n    \t  "
	tmpFile := tmpFileWithContent(content, t)
	defer os.Remove(tmpFile)

	lines, err := ParseEnvFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	expectedLines := []string{
		"foo=bar",
		"baz=quux",
		"_foobar=foobaz",
		"with.dots=working",
		"and_underscore=working too",
	}

	if !reflect.DeepEqual(lines, expectedLines) {
		t.Fatal("lines not equal to expectedLines")
	}
}

// Test ParseEnvFile for an empty file
func TestParseEnvFileEmptyFile(t *testing.T) {
	tmpFile := tmpFileWithContent("", t)
	defer os.Remove(tmpFile)

	lines, err := ParseEnvFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(lines) != 0 {
		t.Fatal("lines not empty; expected empty")
	}
}

// Test ParseEnvFile for a non existent file
func TestParseEnvFileNonExistentFile(t *testing.T) {
	_, err := ParseEnvFile("foo_bar_baz")
	if err == nil {
		t.Fatal("ParseEnvFile succeeded; expected failure")
	}
	if _, ok := err.(*os.PathError); !ok {
		t.Fatalf("Expected a PathError, got [%v]", err)
	}
}

// Test ParseEnvFile for a badly formatted file
func TestParseEnvFileBadlyFormattedFile(t *testing.T) {
	content := `foo=bar
    f   =quux
`

	tmpFile := tmpFileWithContent(content, t)
	defer os.Remove(tmpFile)

	_, err := ParseEnvFile(tmpFile)
	if err == nil {
		t.Fatalf("Expected an ErrBadKey, got nothing")
	}
	if _, ok := err.(ErrBadKey); !ok {
		t.Fatalf("Expected an ErrBadKey, got [%v]", err)
	}
	expectedMessage := "poorly formatted environment: variable 'f   ' contains whitespaces"
	if err.Error() != expectedMessage {
		t.Fatalf("Expected [%v], got [%v]", expectedMessage, err.Error())
	}
}

// Test ParseEnvFile for a file with a line exceeding bufio.MaxScanTokenSize
func TestParseEnvFileLineTooLongFile(t *testing.T) {
	content := strings.Repeat("a", bufio.MaxScanTokenSize+42)
	content = fmt.Sprint("foo=", content)

	tmpFile := tmpFileWithContent(content, t)
	defer os.Remove(tmpFile)

	_, err := ParseEnvFile(tmpFile)
	if err == nil {
		t.Fatal("ParseEnvFile succeeded; expected failure")
	}
}

// ParseEnvFile with a random file, pass through
func TestParseEnvFileRandomFile(t *testing.T) {
	content := `first line
another invalid line`
	tmpFile := tmpFileWithContent(content, t)
	defer os.Remove(tmpFile)

	_, err := ParseEnvFile(tmpFile)
	if err == nil {
		t.Fatalf("Expected an ErrBadKey, got nothing")
	}
	if _, ok := err.(ErrBadKey); !ok {
		t.Fatalf("Expected an ErrBadKey, got [%v]", err)
	}
	expectedMessage := "poorly formatted environment: variable 'first line' contains whitespaces"
	if err.Error() != expectedMessage {
		t.Fatalf("Expected [%v], got [%v]", expectedMessage, err.Error())
	}
}

// ParseEnvFile with environment variable import definitions
func TestParseEnvVariableDefinitionsFile(t *testing.T) {
	content := `# comment=
HOME
UNDEFINED_VAR
`
	tmpFile := tmpFileWithContent(content, t)
	defer os.Remove(tmpFile)

	variables, err := ParseEnvFile(tmpFile)
	if nil != err {
		t.Fatal("There must not be any error")
	}

	if 2 != len(variables) {
		t.Fatal("exactly two variable are imported")
	}

	if "HOME="+os.Getenv("HOME") != variables[0] {
		t.Fatal("the HOME variable is not properly imported as the first variable")
	}

	if "UNDEFINED_VAR" != variables[1] {
		t.Fatal("the UNDEFINED_VAR variable is not properly imported as the second variable")
	}
}

// ParseEnvFile with empty variable name
func TestParseEnvVariableWithNoNameFile(t *testing.T) {
	content := `# comment=
=blank variable names are an error case
`
	tmpFile := tmpFileWithContent(content, t)
	defer os.Remove(tmpFile)

	_, err := ParseEnvFile(tmpFile)
	if nil == err {
		t.Fatal("if a variable has no name parsing an environment file must fail")
	}
}
