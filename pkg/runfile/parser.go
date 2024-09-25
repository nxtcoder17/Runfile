package runfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func parseDotEnv(reader io.Reader) (map[string]string, error) {
	return godotenv.Parse(reader)
}

// parseDotEnv parses the .env file and returns a slice of strings as in os.Environ()
func parseDotEnvFiles(files ...string) (map[string]string, error) {
	results := make(map[string]string)

	for i := range files {
		if !filepath.IsAbs(files[i]) {
			return nil, fmt.Errorf("dotenv file path %s, must be absolute", files[i])
		}

		f, err := os.Open(files[i])
		if err != nil {
			return nil, err
		}
		m, err := parseDotEnv(f)
		if err != nil {
			return nil, err
		}
		f.Close()

		for k, v := range m {
			results[k] = v
		}

	}

	return results, nil
}
