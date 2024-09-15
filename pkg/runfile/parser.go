package runfile

import (
	"os"

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
