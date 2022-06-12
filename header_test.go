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

func Test_headerGroupIndices(t *testing.T) {
	fi, err := excelize.OpenFile(fuseTestFiles[0])
	assert.Nil(t, err)
	h, err := headersFrom(fi)
	assert.Nil(t, err)

	headerGroups, err := headerGroupIndices(h)
	assert.Nil(t, err)

	tests := []struct {
		name   string
		header string
		want   []int
	}{
		{name: "RECORD TYPE", header: "RECORD TYPE", want: []int{1}},
		{name: "Width", header: "Width", want: []int{1336, 2559}},
	}

	for _, test := range tests {
		i, exist := headerGroups[test.header]
		if !exist {
			t.Errorf("expected %s to be stored in header group index map", test.header)
		} else if len(test.want) != len(i) {
			t.Errorf("expected %d indices for header %s, but there were %d", len(test.want), test.header, len(i))
		}

		for i, index := range i {
			if test.want[i] != index {
				t.Errorf("expected index %d for %s to be %d, but got %d", i, test.header, test.want[i], index)
			}
		}
	}
}

func Test_headerIndex(t *testing.T) {
	fi, err := excelize.OpenFile(fuseTestFiles[0])
	assert.Nil(t, err)
	h, err := headersFrom(fi)
	assert.Nil(t, err)

	type args struct {
		headerRow           []string
		keyHeader           string
		otherHeadersInGroup []string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{name: "Width", args: args{headerRow: h, keyHeader: "Width", otherHeadersInGroup: []string{"Trade Item Composition Width UOM"}}, want: 1363, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := headerIndex(tt.args.headerRow, tt.args.keyHeader, tt.args.otherHeadersInGroup)
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
	fi, err := excelize.OpenFile(fuseTestFiles[0])
	assert.Nil(t, err)
	h, err := headersFrom(fi)
	assert.Nil(t, err)

	headerIndices, err := headerGroupIndices(h)
	assert.Nil(t, err)

	type args struct {
		groupIndices   map[string][]int
		headersInGroup []string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{name: "Width", args: args{groupIndices: headerIndices, headersInGroup: []string{"Width", "Trade Item Composition Width UOM"}}, want: 1335, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := headerGroupRootIndex(tt.args.groupIndices, tt.args.headersInGroup)
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
