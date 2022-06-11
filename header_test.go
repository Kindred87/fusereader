package fusereader

import (
	"fmt"
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

func headersFrom(file string) ([]string, error) {
	fi, err := excelize.OpenFile(file)
	if err != nil {
		return nil, fmt.Errorf("error while opening %s: %w", file, err)
	}
	defer fi.Close()

	rows, err := fi.Rows(worksheetFSItem)
	if err != nil {
		return nil, fmt.Errorf("error while initializing row iterator for %s in %s: %w", worksheetFSItem, file, err)
	}

	currRow := 0
	for rows.Next() {
		currRow++

		if currRow >= headerRowMax {
			return nil, fmt.Errorf("could not locate header row within first %d rows in %s", headerRowMax, file)
		}

		r, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("error while reading row %d in %s: %w", currRow, file, err)
		}

		if isHeaderRow(r) {
			return r, nil
		}
	}

	return nil, fmt.Errorf("could not locate header row in %s", file)
}

func Test_headerGroupIndices(t *testing.T) {
	h, err := headersFrom(fuseTestFiles[0])
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
