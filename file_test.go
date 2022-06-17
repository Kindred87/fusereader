package fusereader

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func Test_getFile(t *testing.T) {
	fi, err := excelize.OpenFile(fuseTestFiles[0])
	assert.Nil(t, err)

	type args struct {
		path string
	}
	tests := []struct {
		name     string
		args     args
		want     *excelize.File
		wantErr  bool
		prepFunc func()
	}{
		{name: "Empty cache", args: args{path: fuseTestFiles[0]}, want: nil, wantErr: true},
		{name: "Functional", args: args{path: fuseTestFiles[0]}, want: fi, wantErr: false, prepFunc: func() {
			fileCache = make(map[string]*excelize.File)
			fileCache[fuseTestFiles[0]] = fi
		}},
	}
	for _, tt := range tests {
		if tt.prepFunc != nil {
			tt.prepFunc()
		}

		t.Run(tt.name, func(t *testing.T) {
			got, err := getFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_closeFiles(t *testing.T) {
	_, err := cacheFiles(fuseTestFiles)
	assert.Nil(t, err)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "Basic", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := closeFiles(); (err != nil) != tt.wantErr {
				t.Errorf("closeFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_cacheFiles(t *testing.T) {
	type args struct {
		paths []string
	}
	tests := []struct {
		name    string
		args    args
		want    []*excelize.File
		wantErr bool
	}{
		{name: "Basic", args: args{paths: fuseTestFiles}, wantErr: false},
		{name: "Bogus path", args: args{paths: []string{"foo.xlsx"}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cacheFiles(tt.args.paths)
			if (err != nil) != tt.wantErr {
				t.Errorf("cacheFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cacheFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}
