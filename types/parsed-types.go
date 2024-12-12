package types

type ParsedTask struct {
	Shell       []string            `json:"shell"`
	WorkingDir  string              `json:"workingDir"`
	Env         map[string]string   `json:"environ"`
	Interactive bool                `json:"interactive,omitempty"`
	Commands    []ParsedCommandJson `json:"commands"`
}