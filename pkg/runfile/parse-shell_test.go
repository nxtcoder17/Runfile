package runfile

import (
	"reflect"
	"testing"
)

func TestParsedRunfile_ParseTaskShell(t *testing.T) {
	tests := []struct {
		name     string
		runfile  *ParsedRunfile
		taskName string
		want     Shell
		wantErr  bool
	}{
		{
			name: "1. When task has shell='sh', It should use sh -c",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{
					"test": {
						Shell: "sh",
					},
				},
			},
			taskName: "test",
			want:     []string{"sh", "-c"},
			wantErr:  false,
		},
		{
			name: "2. When task has shell='bash', It should use bash -c",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{
					"build": {
						Shell: "bash",
					},
				},
			},
			taskName: "build",
			want:     []string{"bash", "-c"},
			wantErr:  false,
		},
		{
			name: "3. When task has shell='python', It should use python -c",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{
					"script": {
						Shell: "python",
					},
				},
			},
			taskName: "script",
			want:     []string{"python", "-c"},
			wantErr:  false,
		},
		{
			name: "4. When task has custom shell array, It should use it as-is",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{
					"custom": {
						Shell: []string{"zsh", "-c"},
					},
				},
			},
			taskName: "custom",
			want:     []string{"zsh", "-c"},
			wantErr:  false,
		},
		{
			name: "5. When task has nil shell, It should default to sh -c",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{
					"default": {
						Shell: nil,
					},
				},
			},
			taskName: "default",
			want:     []string{"sh", "-c"},
			wantErr:  false,
		},
		{
			name: "6. When task has invalid shell alias, It should fail",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{
					"invalid": {
						Shell: "invalidshell",
					},
				},
			},
			taskName: "invalid",
			want:     nil,
			wantErr:  true,
		},
		{
			name: "7. When task has invalid shell type (number), It should fail",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{
					"badtype": {
						Shell: 123,
					},
				},
			},
			taskName: "badtype",
			want:     nil,
			wantErr:  true,
		},
		{
			name: "8. When task is not found, It should fail",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{
					"existing": {
						Shell: "sh",
					},
				},
			},
			taskName: "nonexistent",
			want:     nil,
			wantErr:  true,
		},
		{
			name: "9. When task has shell='node', It should use node -e",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{
					"nodejs": {
						Shell: "node",
					},
				},
			},
			taskName: "nodejs",
			want:     []string{"node", "-e"},
			wantErr:  false,
		},
		{
			name: "10. When task has shell='powershell', It should use powershell -Command",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{
					"ps": {
						Shell: "powershell",
					},
				},
			},
			taskName: "ps",
			want:     []string{"powershell", "-Command"},
			wantErr:  false,
		},
		{
			name: "11. When runfile has empty tasks map, It should fail",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{},
			},
			taskName: "any",
			want:     nil,
			wantErr:  true,
		},
		{
			name: "12. When task has complex custom shell array, It should preserve all elements",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{
					"docker": {
						Shell: []string{"docker", "run", "--rm", "alpine", "sh", "-c"},
					},
				},
			},
			taskName: "docker",
			want:     []string{"docker", "run", "--rm", "alpine", "sh", "-c"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.runfile.ParseTaskShell(tt.taskName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsedRunfile.ParseTaskShell() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsedRunfile.ParseTaskShell() = %v, want %v", got, tt.want)
			}
		})
	}
}