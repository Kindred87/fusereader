package fusereader

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func Test_isHeaderRow(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Valid", args: args{s: headerRowPrefix()}, want: true},
		{name: "Empty", args: args{s: []string{""}}, want: false},
		{name: "Partial", args: args{s: headerRowPrefix()[:3]}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isHeaderRow(tt.args.s); got != tt.want {
				t.Errorf("isHeaderRow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_headerRowPrefix(t *testing.T) {
	prefix := headerRowPrefix()

	assert.NotEmpty(t, prefix)
}

func Test_buildHeaderCaches(t *testing.T) {
	cachedFiles, err := cacheFiles(fuseTestFiles)
	defer closeFiles()
	assert.Nil(t, err)

	type args struct {
		files []*excelize.File
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "all files", args: args{files: cachedFiles}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := buildHeaderCaches(tt.args.files...); (err != nil) != tt.wantErr {
				t.Errorf("buildHeaderCaches() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_headersAreShared(t *testing.T) {
	cachedFiles, err := cacheFiles(fuseTestFiles)
	defer closeFiles()
	assert.Nil(t, err)

	fileHeaders, err := assembleHeaders(cachedFiles)
	assert.Nil(t, err)

	bogusHeaders := fileHeaders[0][2000:]
	bogusHeaders = append(bogusHeaders, fileHeaders[0][0:2000]...)

	type args struct {
		headers [][]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "all files", args: args{headers: fileHeaders}, want: true},
		{name: "all files and bogus headers", args: args{headers: append(fileHeaders, bogusHeaders)}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := headersAreShared(tt.args.headers); got != tt.want {
				t.Errorf("headersAreShared() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_headerIndex(t *testing.T) {
	cachedFiles, err := cacheFiles([]string{fuseTestFiles[0]})
	defer closeFiles()
	assert.Nil(t, err)

	err = buildHeaderCaches(cachedFiles...)
	defer removeHeaderCaches()
	assert.Nil(t, err)

	type args struct {
		file                string
		keyHeader           string
		otherHeadersInGroup []string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{name: "Basic", args: args{file: fuseTestFiles[0], keyHeader: "Additional Product Attribute Name", otherHeadersInGroup: []string{"Additional Product Attribute Value"}}, want: 24, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := headerIndex(tt.args.file, tt.args.keyHeader, tt.args.otherHeadersInGroup)
			if (err != nil) != tt.wantErr {
				t.Errorf("headerIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("headerIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_headerGroupRootIndex(t *testing.T) {
	cachedFiles, err := cacheFiles(fuseTestFiles)
	defer closeFiles()
	assert.Nil(t, err)

	err = buildHeaderCaches(cachedFiles[0])
	defer removeHeaderCaches()
	assert.Nil(t, err)

	type args struct {
		file           string
		headersInGroup []string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{name: "Width", args: args{file: fuseTestFiles[0], headersInGroup: []string{"Width", "Trade Item Composition Width UOM"}}, want: 1335, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := headerGroupRootIndex(tt.args.file, tt.args.headersInGroup)
			if (err != nil) != tt.wantErr {
				t.Errorf("headerGroupRootIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("headerGroupRootIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}
