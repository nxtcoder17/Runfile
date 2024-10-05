package runfile

import (
	"os"
	"path/filepath"

	fn "github.com/nxtcoder17/runfile/pkg/functions"
	"github.com/nxtcoder17/runfile/pkg/runfile/errors"
	"sigs.k8s.io/yaml"
)

type attrs struct {
	RunfilePath string
}

type Runfile struct {
	attrs attrs

	Version  string                 `json:"version,omitempty"`
	Includes map[string]IncludeSpec `json:"includes"`
	Tasks    map[string]Task        `json:"tasks"`
}

type IncludeSpec struct {
	Runfile string `json:"runfile"`
	Dir     string `json:"dir,omitempty"`
}

type ParsedIncludeSpec struct {
	Runfile *Runfile
}

func Parse(file string) (*Runfile, errors.Message) {
	var runfile Runfile
	f, err := os.ReadFile(file)
	if err != nil {
		return &runfile, errors.RunfileReadFailed.WithErr(err).WithMetadata("file", file)
	}
	if err = yaml.Unmarshal(f, &runfile); err != nil {
		return nil, errors.RunfileParsingFailed.WithErr(err).WithMetadata("file", file)
	}

	runfile.attrs.RunfilePath = fn.Must(filepath.Abs(file))
	return &runfile, nil
}

func (rf *Runfile) ParseIncludes() (map[string]ParsedIncludeSpec, errors.Message) {
	m := make(map[string]ParsedIncludeSpec, len(rf.Includes))
	for k, v := range rf.Includes {
		r, err := Parse(v.Runfile)
		if err != nil {
			return nil, err.WithMetadata("includes", v.Runfile)
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
