package runfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func parseDotEnv(reader io.Reader) (map[string]string, *Error) {
	m, err := godotenv.Parse(reader)
	if err != nil {
		return nil, DotEnvParsingFailed.WithErr(err)
	}
	return m, nil
}

// parseDotEnv parses the .env file and returns a slice of strings as in os.Environ()

func parseDotEnvFiles(files ...string) (map[string]string, *Error) {
	results := make(map[string]string)

	for i := range files {
		if !filepath.IsAbs(files[i]) {
			return nil, DotEnvInvalid.WithErr(fmt.Errorf("dotenv file paths must be absolute")).WithMetadata("dotenv", files[i])
		}

		f, err := os.Open(files[i])
		if err != nil {
			return nil, DotEnvInvalid.WithErr(err).WithMetadata("dotenv", files[i])
		}

		m, err2 := parseDotEnv(f)
		if err2 != nil {
			return nil, err2.WithMetadata("dotenv", files[i])
		}
		f.Close()

		for k, v := range m {
			results[k] = v
		}

	}

	return results, nil
}

func ParseIncludes(rf *Runfile) (map[string]ParsedIncludeSpec, *Error) {
	m := make(map[string]ParsedIncludeSpec, len(rf.Includes))
	for k, v := range rf.Includes {
		r, err := Parse(v.Runfile)
		if err != nil {
			return nil, err
		}

		for it := range r.Tasks {
			if v.Dir != "" {
				nt := r.Tasks[it]
				nt.Dir = &v.Dir
				r.Tasks[it] = nt
			}
		}

		m[k] = ParsedIncludeSpec{Runfile: r}
	}

	return m, nil
}
