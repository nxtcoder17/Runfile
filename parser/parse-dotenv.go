package parser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/nxtcoder17/runfile/errors"
)

func parseDotEnv(reader io.Reader) (map[string]string, error) {
	m, err := godotenv.Parse(reader)
	if err != nil {
		return nil, errors.ErrParseDotEnv.Wrap(err)
	}
	return m, nil
}

// parseDotEnv parses the .env file and returns a slice of strings as in os.Environ()

func parseDotEnvFiles(files ...string) (map[string]string, error) {
	results := make(map[string]string)

	for i := range files {
		if !filepath.IsAbs(files[i]) {
			return nil, errors.ErrInvalidDotEnv.Wrap(fmt.Errorf("dotenv file paths must be absolute")).KV("dotenv", files[i])
		}

		f, err := os.Open(files[i])
		if err != nil {
			return nil, errors.ErrInvalidDotEnv.Wrap(err).KV("dotenv", files[i])
		}

		m, err2 := parseDotEnv(f)
		if err2 != nil {
			return nil, errors.ErrInvalidDotEnv.KV("dotenv", files[i])
		}
		f.Close()

		for k, v := range m {
			results[k] = v
		}
	}

	return results, nil
}
