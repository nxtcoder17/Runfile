package runfile

import (
	"os"
	"path/filepath"

	fn "github.com/nxtcoder17/runfile/pkg/functions"
	"sigs.k8s.io/yaml"
)

type attrs struct {
	RunfilePath string
}

type Runfile struct {
	attrs attrs

	Version string
	Tasks   map[string]Task `json:"tasks"`
}

func Parse(file string) (*Runfile, error) {
	var runfile Runfile
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
