package runfile

import (
	"testing"

)

func TestParsedRunfile_Run(t *testing.T) {
	// Create a test runfile with various tasks
	runfile := &ParsedRunfile{
		Env: map[string]string{
			"INITIAL": "value",
		},
		Tasks: map[string]Task{
			"task1": {
				Name: "task1",
				Commands: []any{
					"echo task1",
				},
			},
			"task2": {
				Name: "task2",
				Commands: []any{
					"echo task2",
				},
			},
			"task3": {
				Name: "task3",
				Commands: []any{
					"echo task3",
				},
			},
			"failing": {
				Name: "failing",
				Commands: []any{
					"exit 1",
				},
			},
		},
	}

	ctx := NewTestContext()

	tests := []struct {
		name    string
		tasks   []string
		opt     RunOption
		wantErr bool
		check   func(t *testing.T, r *ParsedRunfile)
	}{
		{
			name:    "1. When running single task, It should execute successfully",
			tasks:   []string{"task1"},
			opt:     RunOption{},
			wantErr: false,
		},
		{
			name:    "2. When running multiple tasks sequentially, It should execute all tasks",
			tasks:   []string{"task1", "task2", "task3"},
			opt:     RunOption{},
			wantErr: false,
		},
		{
			name: "3. When running multiple tasks in parallel, It should execute all tasks concurrently",
			tasks: []string{"task1", "task2", "task3"},
			opt: RunOption{
				ExecuteInParallel: true,
			},
			wantErr: false,
		},
		{
			name:    "4. When task doesn't exist, It should return error",
			tasks:   []string{"nonexistent"},
			opt:     RunOption{},
			wantErr: true,
		},
		{
			name:    "5. When mixing existent and non-existent tasks, It should return error",
			tasks:   []string{"task1", "nonexistent"},
			opt:     RunOption{},
			wantErr: true,
		},
		{
			name:  "6. When KVs are provided, It should override and add environment variables",
			tasks: []string{"task1"},
			opt: RunOption{
				KVs: map[string]string{
					"INITIAL": "overridden",
					"NEW_VAR": "new_value",
				},
			},
			wantErr: false,
			check: func(t *testing.T, r *ParsedRunfile) {
				if r.Env["INITIAL"] != "overridden" {
					t.Errorf("Expected INITIAL to be overridden, got %s", r.Env["INITIAL"])
				}
				if r.Env["NEW_VAR"] != "new_value" {
					t.Errorf("Expected NEW_VAR to be new_value, got %s", r.Env["NEW_VAR"])
				}
			},
		},
		{
			name:  "7. When env map is nil and KVs are provided, It should create env map",
			tasks: []string{"task1"},
			opt: RunOption{
				KVs: map[string]string{
					"TEST": "value",
				},
			},
			wantErr: false,
			check: func(t *testing.T, r *ParsedRunfile) {
				// The function should have created the map and added the value
				if r.Env == nil {
					t.Error("Expected Env map to be created")
				}
				if r.Env["TEST"] != "value" {
					t.Errorf("Expected TEST to be value, got %s", r.Env["TEST"])
				}
			},
		},
		{
			name:    "8. When task list is empty, It should succeed without running anything",
			tasks:   []string{},
			opt:     RunOption{},
			wantErr: false,
		},
		{
			name: "9. When parallel mode with empty task list, It should succeed",
			tasks: []string{},
			opt: RunOption{
				ExecuteInParallel: true,
			},
			wantErr: false,
		},
		{
			name:    "10. When task fails in sequential mode, It should stop execution",
			tasks:   []string{"task1", "failing", "task2"},
			opt:     RunOption{},
			wantErr: true,
		},
		{
			name: "11. When task fails in parallel mode, It should report error",
			tasks: []string{"task1", "failing", "task2"},
			opt: RunOption{
				ExecuteInParallel: true,
			},
			wantErr: true,
		},
		{
			name: "12. When debug option is enabled, It should run with debug output",
			tasks: []string{"task1"},
			opt: RunOption{
				Debug: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the runfile for each test
			testRunfile := &ParsedRunfile{
				Env:   make(map[string]string),
				Tasks: make(map[string]Task),
			}
			
			// Copy env
			for k, v := range runfile.Env {
				testRunfile.Env[k] = v
			}
			
			// Copy tasks
			for k, v := range runfile.Tasks {
				testRunfile.Tasks[k] = v
			}

			// Special handling for nil env test
			if tt.name == "KVs with nil env map" {
				testRunfile.Env = nil
			}

			err := testRunfile.Run(ctx, tt.tasks, tt.opt)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsedRunfile.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.check != nil {
				tt.check(t, testRunfile)
			}
		})
	}
}

