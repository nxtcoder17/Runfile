package runfile

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

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
		return &runfile, err
	}
	return &runfile, nil
}

// parseDotEnv parses the .env file and returns a slice of strings as in os.Environ()
func parseDotEnv(files ...string) ([]string, error) {
	results := make([]string, 0, 5)

	for i := range files {
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
