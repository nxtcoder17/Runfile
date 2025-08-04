package runfile

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_ParseDotEnvFile(t *testing.T) {
	type args struct {
		reader io.Reader
	}

	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "1. When key=, It should parse as empty string",
			args: args{
				reader: bytes.NewBuffer([]byte(`key=`)),
			},
			want: map[string]string{
				"key": "",
			},
			wantErr: false,
		},
		{
			name: "2. When key=1, It should parse as string",
			args: args{
				reader: bytes.NewBuffer([]byte(`key=1`)),
			},
			want: map[string]string{
				"key": "1",
			},
			wantErr: false,
		},
		{
			name: "3. When key=one, It should parse correctly",
			args: args{
				reader: bytes.NewBuffer([]byte(`key=one`)),
			},
			want: map[string]string{
				"key": "one",
			},
			wantErr: false,
		},
		{
			name: "4. When key='one', It should strip quotes",
			args: args{
				reader: bytes.NewBuffer([]byte(`key='one'`)),
			},
			want: map[string]string{
				"key": "one",
			},
			wantErr: false,
		},
		{
			name: `5. When key='o"ne', It should preserve inner quotes`,
			args: args{
				reader: bytes.NewBuffer([]byte(`key='o"ne'`)),
			},
			want: map[string]string{
				"key": `o"ne`,
			},
			wantErr: false,
		},
		{
			name: `6. When key="one", It should strip quotes`,
			args: args{
				reader: bytes.NewBuffer([]byte(`key="one"`)),
			},
			want: map[string]string{
				"key": `one`,
			},
			wantErr: false,
		},
		{
			name: `7. When key=sample==, It should preserve all equals signs`,
			args: args{
				reader: bytes.NewBuffer([]byte(`key=sample==`)),
			},
			want: map[string]string{
				"key": `sample==`,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDotEnv(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDotEnvFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDotEnvFile()\n\t got: %#v,\n\twant: %#v", got, tt.want)
			}
		})
	}
}

func TestParseDotEnvFiles(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	// Create test .env files
	envFile1 := filepath.Join(tmpDir, "env1.env")
	if err := os.WriteFile(envFile1, []byte("KEY1=value1\nKEY2=value2"), 0o644); err != nil {
		t.Fatal(err)
	}

	envFile2 := filepath.Join(tmpDir, "env2.env")
	if err := os.WriteFile(envFile2, []byte("KEY2=overridden\nKEY3=value3"), 0o644); err != nil {
		t.Fatal(err)
	}

	emptyFile := filepath.Join(tmpDir, "empty.env")
	if err := os.WriteFile(emptyFile, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	invalidFile := filepath.Join(tmpDir, "invalid.env")
	if err := os.WriteFile(invalidFile, []byte("INVALID LINE WITHOUT EQUALS"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		files   []string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "1. When single file is provided, It should parse all key-value pairs",
			files: []string{envFile1},
			want: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			wantErr: false,
		},
		{
			name:  "2. When multiple files are provided, It should merge with later files overriding earlier",
			files: []string{envFile1, envFile2},
			want: map[string]string{
				"KEY1": "value1",
				"KEY2": "overridden",
				"KEY3": "value3",
			},
			wantErr: false,
		},
		{
			name:    "3. When empty file list is provided, It should return empty map",
			files:   []string{},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "4. When file is empty, It should return empty map",
			files:   []string{emptyFile},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "5. When non-existent file is provided, It should return error",
			files:   []string{filepath.Join(tmpDir, "nonexistent.env")},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "6. When relative path is provided, It should fail",
			files:   []string{"relative.env"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "7. When mix of valid and relative paths is provided, It should fail",
			files:   []string{envFile1, "relative.env"},
			want:    nil,
			wantErr: true,
		},
		{
			name:  "8. When files are provided in different order, It should respect the order",
			files: []string{envFile2, envFile1},
			want: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2", // envFile1 overrides envFile2
				"KEY3": "value3",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDotEnvFiles(tt.files...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDotEnvFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDotEnvFiles()\n\tgot:  %#v\n\twant: %#v", got, tt.want)
			}
		})
	}
}
