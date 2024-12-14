package parser

import (
	"github.com/nxtcoder17/runfile/errors"
	"github.com/nxtcoder17/runfile/types"
)

func parseIncludes(includes map[string]types.IncludeSpec) (map[string]*types.ParsedRunfile, error) {
	m := make(map[string]*types.ParsedRunfile, len(includes))
	for k, v := range includes {
		r, err := parseRunfileFromFile(v.Runfile)
		if err != nil {
			return nil, errors.ErrParseIncludes.Wrap(err).KV("include", v.Runfile)
		}

		for it := range r.Tasks {
			if v.Dir != "" {
				nt := r.Tasks[it]
				nt.Dir = &v.Dir
				r.Tasks[it] = nt
			}
		}

		m[k] = r
	}

	return m, nil
}
