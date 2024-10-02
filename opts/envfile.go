package opts

import (
	"fmt"
	"os"
)

// ParseEnvFile reads a file with environment variables enumerated by lines
//
// “Environment variable names used by the utilities in the Shell and
// Utilities volume of IEEE Std 1003.1-2001 consist solely of uppercase
// letters, digits, and the '_' (underscore) from the characters defined in
// Portable Character Set and do not begin with a digit. *But*, other
// characters may be permitted by an implementation; applications shall
// tolerate the presence of such names.”
// -- http://pubs.opengroup.org/onlinepubs/009695399/basedefs/xbd_chap08.html
//
// As of #16585, it's up to application inside docker to validate or not
// environment variables, that's why we just strip leading whitespace and
// nothing more.
func ParseEnvFile(filename string) ([]string, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return []string{}, err
	}
	out, err := parseKeyValueFile(fh, os.LookupEnv)
	_ = fh.Close()
	if err != nil {
		return []string{}, fmt.Errorf("invalid env file (%s): %v", filename, err)
	}
	return out, nil
}
