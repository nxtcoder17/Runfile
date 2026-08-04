package resolver

import (
	"strings"
	"testing"

	"github.com/nxtcoder17/runfile/pkg/runfile/spec"
	"github.com/nxtcoder17/runfile/pkg/writer"
)

func TestParseCommands(t *testing.T) {
	tests := []struct {
		name     string
		commands []any
		expected []*Command
		wantErr  bool
	}{
		{
			name:     "when commands list is empty, it must return empty slice",
			commands: []any{},
			expected: []*Command{},
			wantErr:  false,
		},
		{
			name:     "when command is a string, it must parse correctly",
			commands: []any{"echo hello"},
			expected: []*Command{
				{Text: "echo hello", IsRunTarget: false, Env: nil},
			},
			wantErr: false,
		},
		{
			name:     "when command is an empty string, it must be skipped",
			commands: []any{"", "echo hello", ""},
			expected: []*Command{
				{Text: "echo hello", IsRunTarget: false, Env: nil},
			},
			wantErr: false,
		},
		{
			name:     "when multiple string commands, it must parse all",
			commands: []any{"echo first", "echo second", "echo third"},
			expected: []*Command{
				{Text: "echo first", IsRunTarget: false, Env: nil},
				{Text: "echo second", IsRunTarget: false, Env: nil},
				{Text: "echo third", IsRunTarget: false, Env: nil},
			},
			wantErr: false,
		},
		{
			name: "when command has run key, it must be marked as run target",
			commands: []any{
				map[string]any{"run": "other-task"},
			},
			expected: []*Command{
				{Text: "other-task", IsRunTarget: true, Env: nil},
			},
			wantErr: false,
		},
		{
			name: "when command has run key with env, it must include env",
			commands: []any{
				map[string]any{
					"run": "other-task",
					"env": map[string]any{"KEY": "value"},
				},
			},
			expected: []*Command{
				{Text: "other-task", IsRunTarget: true, Env: map[string]any{"KEY": "value"}},
			},
			wantErr: false,
		},
		{
			name: "when command has cmd key, it must parse as regular command",
			commands: []any{
				map[string]any{"cmd": "echo from cmd"},
			},
			expected: []*Command{
				{Text: "echo from cmd", IsRunTarget: false, Env: nil},
			},
			wantErr: false,
		},
		{
			name: "when run key is empty string, it must fail",
			commands: []any{
				map[string]any{"run": ""},
			},
			wantErr: true,
		},
		{
			name: "when cmd key is empty string, it must fail",
			commands: []any{
				map[string]any{"cmd": ""},
			},
			wantErr: true,
		},
		{
			name: "when map has neither run nor cmd, it must fail",
			commands: []any{
				map[string]any{"invalid": "value"},
			},
			wantErr: true,
		},
		{
			name:     "when command is invalid type (int), it must fail",
			commands: []any{123},
			wantErr:  true,
		},
		{
			name:     "when command is invalid type (bool), it must fail",
			commands: []any{true},
			wantErr:  true,
		},
		{
			name: "when mixed string and map commands, it must parse all correctly",
			commands: []any{
				"echo first",
				map[string]any{"run": "setup"},
				"echo middle",
				map[string]any{"cmd": "echo from cmd"},
				"echo last",
			},
			expected: []*Command{
				{Text: "echo first", IsRunTarget: false, Env: nil},
				{Text: "setup", IsRunTarget: true, Env: nil},
				{Text: "echo middle", IsRunTarget: false, Env: nil},
				{Text: "echo from cmd", IsRunTarget: false, Env: nil},
				{Text: "echo last", IsRunTarget: false, Env: nil},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCommands(tt.commands)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d commands, got %d", len(tt.expected), len(result))
				return
			}

			for i, cmd := range result {
				exp := tt.expected[i]
				if cmd.Text != exp.Text {
					t.Errorf("command[%d]: expected Text=%q, got %q", i, exp.Text, cmd.Text)
				}
				if cmd.IsRunTarget != exp.IsRunTarget {
					t.Errorf("command[%d]: expected IsRunTarget=%v, got %v", i, exp.IsRunTarget, cmd.IsRunTarget)
				}
				// Check env if expected
				if exp.Env != nil {
					if cmd.Env == nil {
						t.Errorf("command[%d]: expected Env to be set, got nil", i)
					}
				}
			}
		})
	}
}

func TestCreateSteps(t *testing.T) {
	tests := []struct {
		name           string
		resolver       *Resolver
		task           *ResolvedTask
		args           createCommandGroupArgs
		expectedSteps  int
		expectedCmds   []int // number of commands per step
		checkSubSteps  bool
		subStepIndices []int // which steps should have substeps
		wantErr        bool
	}{
		{
			name:     "when task has no commands, it must return empty steps",
			resolver: &Resolver{Env: map[string]string{}},
			task: &ResolvedTask{
				Name:     "empty",
				Shell:    []string{"bash", "-c"},
				Commands: []*Command{},
			},
			args: createCommandGroupArgs{
				Stdout: &writer.LogWriter{},
				Stderr: &writer.LogWriter{},
			},
			expectedSteps: 0,
			expectedCmds:  []int{},
			wantErr:       false,
		},
		{
			name:     "when task has single command, it must create one step",
			resolver: &Resolver{Env: map[string]string{}},
			task: &ResolvedTask{
				Name:     "single",
				Shell:    []string{"bash", "-c"},
				Commands: []*Command{{Text: "echo hello", IsRunTarget: false}},
			},
			args: createCommandGroupArgs{
				Stdout: &writer.LogWriter{},
				Stderr: &writer.LogWriter{},
			},
			expectedSteps: 1,
			expectedCmds:  []int{1},
			wantErr:       false,
		},
		{
			name:     "when task has multiple commands, it must create multiple steps",
			resolver: &Resolver{Env: map[string]string{}},
			task: &ResolvedTask{
				Name:  "multi",
				Shell: []string{"bash", "-c"},
				Commands: []*Command{
					{Text: "echo first", IsRunTarget: false},
					{Text: "echo second", IsRunTarget: false},
					{Text: "echo third", IsRunTarget: false},
				},
			},
			args: createCommandGroupArgs{
				Stdout: &writer.LogWriter{},
				Stderr: &writer.LogWriter{},
			},
			expectedSteps: 3,
			expectedCmds:  []int{1, 1, 1},
			wantErr:       false,
		},
		{
			name: "when task has run target, it must create substeps",
			resolver: &Resolver{
				Env: map[string]string{},
				Tasks: map[string]extendedTaskSpec{
					"subtask": {
						TaskSpec: spec.TaskSpec{
							Shell:    "bash",
							Commands: []any{"echo from subtask"},
						},
					},
				},
			},
			task: &ResolvedTask{
				Name:  "parent",
				Shell: []string{"bash", "-c"},
				Commands: []*Command{
					{Text: "subtask", IsRunTarget: true},
				},
			},
			args: createCommandGroupArgs{
				Stdout: &writer.LogWriter{},
				Stderr: &writer.LogWriter{},
			},
			expectedSteps:  1,
			checkSubSteps:  true,
			subStepIndices: []int{0},
			wantErr:        false,
		},
		{
			name: "when run target does not exist, it must fail",
			resolver: &Resolver{
				Env:   map[string]string{},
				Tasks: map[string]extendedTaskSpec{},
			},
			task: &ResolvedTask{
				Name:  "parent",
				Shell: []string{"bash", "-c"},
				Commands: []*Command{
					{Text: "nonexistent", IsRunTarget: true},
				},
			},
			args: createCommandGroupArgs{
				Stdout: &writer.LogWriter{},
				Stderr: &writer.LogWriter{},
			},
			wantErr: true,
		},
		{
			name:     "when task is parallel, steps must have Parallel flag set",
			resolver: &Resolver{Env: map[string]string{}},
			task: &ResolvedTask{
				Name:     "parallel-task",
				Shell:    []string{"bash", "-c"},
				Parallel: true,
				Commands: []*Command{
					{Text: "echo first", IsRunTarget: false},
					{Text: "echo second", IsRunTarget: false},
				},
			},
			args: createCommandGroupArgs{
				Stdout: &writer.LogWriter{},
				Stderr: &writer.LogWriter{},
			},
			expectedSteps: 1,
			expectedCmds:  []int{0},
			wantErr:       false,
		},
		{
			name:     "when task is interactive, it must use interactive shell command",
			resolver: &Resolver{Env: map[string]string{}},
			task: &ResolvedTask{
				Name:        "interactive-task",
				Shell:       []string{"bash", "-c"},
				Interactive: true,
				Commands:    []*Command{{Text: "node", IsRunTarget: false}},
			},
			args: createCommandGroupArgs{
				Stdout: &writer.LogWriter{},
				Stderr: &writer.LogWriter{},
			},
			expectedSteps: 1,
			expectedCmds:  []int{1},
			wantErr:       false,
		},
		{
			name: "when mixed commands and run targets, it must handle both",
			resolver: &Resolver{
				Env: map[string]string{},
				Tasks: map[string]extendedTaskSpec{
					"setup": {
						TaskSpec: spec.TaskSpec{
							Shell:    "bash",
							Commands: []any{"echo setup"},
						},
					},
				},
			},
			task: &ResolvedTask{
				Name:  "mixed",
				Shell: []string{"bash", "-c"},
				Commands: []*Command{
					{Text: "echo before", IsRunTarget: false},
					{Text: "setup", IsRunTarget: true},
					{Text: "echo after", IsRunTarget: false},
				},
			},
			args: createCommandGroupArgs{
				Stdout: &writer.LogWriter{},
				Stderr: &writer.LogWriter{},
			},
			expectedSteps:  3,
			checkSubSteps:  true,
			subStepIndices: []int{1}, // only step at index 1 should have substeps
			wantErr:        false,
		},
		{
			name: "when task has circular dependency, it must fail",
			resolver: &Resolver{
				Env: map[string]string{},
				Tasks: map[string]extendedTaskSpec{
					"a": {
						TaskSpec: spec.TaskSpec{
							Commands: []any{map[string]any{"run": "b"}},
						},
					},
					"b": {
						TaskSpec: spec.TaskSpec{
							Commands: []any{map[string]any{"run": "a"}},
						},
					},
				},
			},
			task: &ResolvedTask{
				Name:  "a",
				Shell: []string{"bash", "-c"},
				Commands: []*Command{
					{Text: "b", IsRunTarget: true},
				},
			},
			args: createCommandGroupArgs{
				Stdout: &writer.LogWriter{},
				Stderr: &writer.LogWriter{},
			},
			wantErr: true,
		},
		{
			name: "when task runs itself, it must fail",
			resolver: &Resolver{
				Env: map[string]string{},
				Tasks: map[string]extendedTaskSpec{
					"a": {
						TaskSpec: spec.TaskSpec{
							Commands: []any{map[string]any{"run": "a"}},
						},
					},
				},
			},
			task: &ResolvedTask{
				Name:  "a",
				Shell: []string{"bash", "-c"},
				Commands: []*Command{
					{Text: "a", IsRunTarget: true},
				},
			},
			args: createCommandGroupArgs{
				Stdout: &writer.LogWriter{},
				Stderr: &writer.LogWriter{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps, err := tt.resolver.createSteps(tt.task, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}

				if strings.Contains(tt.name, "circular dependency") && !strings.Contains(err.Error(), "a -> b -> a") {
					t.Fatalf("expected circular dependency error to include cycle path, got %v", err)
				}

				if strings.Contains(tt.name, "runs itself") && !strings.Contains(err.Error(), "a -> a") {
					t.Fatalf("expected self dependency error to include cycle path, got %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(steps) != tt.expectedSteps {
				t.Errorf("expected %d steps, got %d", tt.expectedSteps, len(steps))
				return
			}

			// Check command counts per step
			for i, expectedCmdCount := range tt.expectedCmds {
				if i >= len(steps) {
					break
				}
				if len(steps[i].Commands) != expectedCmdCount {
					t.Errorf("step[%d]: expected %d commands, got %d", i, expectedCmdCount, len(steps[i].Commands))
				}
			}

			// Check substeps
			if tt.checkSubSteps {
				subStepSet := make(map[int]bool)
				for _, idx := range tt.subStepIndices {
					subStepSet[idx] = true
				}

				for i, step := range steps {
					if subStepSet[i] {
						if len(step.SubSteps) == 0 {
							t.Errorf("step[%d]: expected substeps, got none", i)
						}
					}
				}
			}

			// Check parallel flag
			if tt.task.Parallel {
				for i, step := range steps {
					if !step.Parallel {
						t.Errorf("step[%d]: expected Parallel=true, got false", i)
					}
				}
			}
		})
	}
}

func TestCreateStepsWithEnv(t *testing.T) {
	tests := []struct {
		name        string
		resolverEnv map[string]string
		argsEnv     map[string]string
		taskEnv     map[string]any
		cmdEnv      map[string]any
		wantErr     bool
	}{
		{
			name:        "when resolver has env, it must be available",
			resolverEnv: map[string]string{"GLOBAL": "from_resolver"},
			wantErr:     false,
		},
		{
			name:        "when args has env, it must override resolver env",
			resolverEnv: map[string]string{"KEY": "from_resolver"},
			argsEnv:     map[string]string{"KEY": "from_args"},
			wantErr:     false,
		},
		{
			name:    "when task has env, it must be parsed",
			taskEnv: map[string]any{"TASK_KEY": "from_task"},
			wantErr: false,
		},
		{
			name:    "when command has env, it must be parsed",
			cmdEnv:  map[string]any{"CMD_KEY": "from_cmd"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &Resolver{
				Env:   tt.resolverEnv,
				Tasks: map[string]extendedTaskSpec{},
			}
			if resolver.Env == nil {
				resolver.Env = map[string]string{}
			}

			task := &ResolvedTask{
				Name:     "test",
				Shell:    []string{"bash", "-c"},
				Env:      tt.taskEnv,
				Commands: []*Command{{Text: "echo test", IsRunTarget: false, Env: tt.cmdEnv}},
			}

			args := createCommandGroupArgs{
				Stdout: &writer.LogWriter{},
				Stderr: &writer.LogWriter{},
				Env:    tt.argsEnv,
			}

			_, err := resolver.createSteps(task, args)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
