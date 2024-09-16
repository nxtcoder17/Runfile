package runfile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	fn "github.com/nxtcoder17/runfile/pkg/functions"
	"sigs.k8s.io/yaml"
)

func ParseRunFile(file string) (*RunFile, error) {
	var runfile RunFile
	f, err := os.ReadFile(file)
	if err != nil {
		return &runfile, err
	}
	err = yaml.Unmarshal(f, &runfile)
	if err != nil {
		return nil, err
	}

	runfile.attrs.RunfilePath = fn.Must(filepath.Abs(file))
	return &runfile, nil
}

// parseDotEnv parses the .env file and returns a slice of strings as in os.Environ()
func parseDotEnv(files ...string) ([]string, error) {
	results := make([]string, 0, 5)

	for i := range files {
		if !filepath.IsAbs(files[i]) {
			return nil, fmt.Errorf("dotenv file path %s, must be absolute", files[i])
		}

		f, err := os.Open(files[i])
		if err != nil {
			return nil, err
		}

		s := bufio.NewScanner(f)
		for s.Scan() {
			s2 := strings.SplitN(s.Text(), "=", 2)
			if len(s2) != 2 {
				continue
			}
			s, _ := strconv.Unquote(string(s2[1]))

			// os.Setenv(s2[0], s2[1])
			os.Setenv(s2[0], s)
			results = append(results, s2[0])
		}
	}

	for i := range results {
		v := os.Getenv(results[i])
		results[i] = fmt.Sprintf("%s=%v", results[i], v)
	}

	return results, nil
}
