package runfile

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func Test_parseDotEnvFile(t *testing.T) {
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
			name: "key=",
			args: args{
				reader: bytes.NewBuffer([]byte(`key=`)),
			},
			want: map[string]string{
				"key": "",
			},
			wantErr: false,
		},
		{
			name: "key=1",
			args: args{
				reader: bytes.NewBuffer([]byte(`key=1`)),
			},
			want: map[string]string{
				"key": "1",
			},
			wantErr: false,
		},
		{
			name: "key=one",
			args: args{
				reader: bytes.NewBuffer([]byte(`key=one`)),
			},
			want: map[string]string{
				"key": "one",
			},
			wantErr: false,
		},
		{
			name: "key='one'",
			args: args{
				reader: bytes.NewBuffer([]byte(`key='one'`)),
			},
			want: map[string]string{
				"key": "one",
			},
			wantErr: false,
		},
		{
			name: `key='o"ne'`,
			args: args{
				reader: bytes.NewBuffer([]byte(`key='o"ne'`)),
			},
			want: map[string]string{
				"key": `o"ne`,
			},
			wantErr: false,
		},
		{
			name: `key="one"`,
			args: args{
				reader: bytes.NewBuffer([]byte(`key="one"`)),
			},
			want: map[string]string{
				"key": `one`,
			},
			wantErr: false,
		},
		{
			name: `key=sample==`,
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
