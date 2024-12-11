package parser

import (
	"bytes"
	"io"
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
			name: "1. key=",
			args: args{
				reader: bytes.NewBuffer([]byte(`key=`)),
			},
			want: map[string]string{
				"key": "",
			},
			wantErr: false,
		},
		{
			name: "2. key=1",
			args: args{
				reader: bytes.NewBuffer([]byte(`key=1`)),
			},
			want: map[string]string{
				"key": "1",
			},
			wantErr: false,
		},
		{
			name: "3. key=one",
			args: args{
				reader: bytes.NewBuffer([]byte(`key=one`)),
			},
			want: map[string]string{
				"key": "one",
			},
			wantErr: false,
		},
		{
			name: "4. key='one'",
			args: args{
				reader: bytes.NewBuffer([]byte(`key='one'`)),
			},
			want: map[string]string{
				"key": "one",
			},
			wantErr: false,
		},
		{
			name: `5. key='o"ne'`,
			args: args{
				reader: bytes.NewBuffer([]byte(`key='o"ne'`)),
			},
			want: map[string]string{
				"key": `o"ne`,
			},
			wantErr: false,
		},
		{
			name: `6. key="one"`,
			args: args{
				reader: bytes.NewBuffer([]byte(`key="one"`)),
			},
			want: map[string]string{
				"key": `one`,
			},
			wantErr: false,
		},
		{
			name: `7. key=sample==`,
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
