package runfile

import (
	"os"
	"path/filepath"

	fn "github.com/nxtcoder17/runfile/pkg/functions"
	"sigs.k8s.io/yaml"
)

type attrs struct {
	RootRunfilePath string
	RunfilePath     string
}

type Runfile struct {
	attrs attrs

	Version  string                 `json:"version,omitempty"`
	Includes map[string]IncludeSpec `json:"includes"`
	Env      EnvVar                 `json:"env,omitempty"`
	DotEnv   []string               `json:"dotEnv,omitempty"`
	Tasks    map[string]Task        `json:"tasks"`
}

type IncludeSpec struct {
	Runfile string `json:"runfile"`
	Dir     string `json:"dir,omitempty"`
}

type ParsedIncludeSpec struct {
	Runfile *Runfile
}

func Parse(file string) (*Runfile, *Error) {
	var runfile Runfile
	f, err := os.ReadFile(file)
	if err != nil {
		return &runfile, RunfileReadFailed.WithErr(err).WithMetadata("file", file)
	}
	if err = yaml.Unmarshal(f, &runfile); err != nil {
		return nil, RunfileParsingFailed.WithErr(err).WithMetadata("file", file)
	}

	runfile.attrs.RunfilePath = fn.Must(filepath.Abs(file))
	runfile.attrs.RootRunfilePath = runfile.attrs.RunfilePath
	return &runfile, nil
}

func (rf *Runfile) parse(file string) (*Runfile, *Error) {
	r, err := Parse(file)
	if err != nil {
		return nil, err
	}
	r.attrs.RootRunfilePath = rf.attrs.RunfilePath
	return r, nil
}

func (rf *Runfile) ParseIncludes() (map[string]ParsedIncludeSpec, *Error) {
	m := make(map[string]ParsedIncludeSpec, len(rf.Includes))
	for k, v := range rf.Includes {
		r, err := rf.parse(v.Runfile)
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
