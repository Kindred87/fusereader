package fusereader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocationOf(t *testing.T) {
	file, row, err := itemLocation("00046015128797", fuseTestFiles)

	assert.Nil(t, err)

	assert.Equal(t, "testdata/fuse01.xlsx", file)
	assert.Equal(t, 25, row)
}

func Test_itemLocation(t *testing.T) {
	type args struct {
		itemID string
		files  []string
		opts   []Option
	}
	tests := []struct {
		name     string
		args     args
		wantFile string
		wantRow  int
		wantErr  bool
	}{
		{name: "invalid item ID", args: args{itemID: "9999999999999", files: fuseTestFiles}, wantErr: true},
		{name: "10077661153618", args: args{itemID: "10077661153618", files: fuseTestFiles}, wantFile: fuseTestFiles[0], wantRow: 5348, wantErr: false},
		{name: "near-match", args: args{itemID: "000460151672841", files: fuseTestFiles}, wantErr: true},
		{name: "00077661142141", args: args{itemID: "00077661142141", files: fuseTestFiles}, wantFile: fuseTestFiles[1], wantRow: 9236, wantErr: false},
		{name: "00077661004128", args: args{itemID: "00077661004128", files: fuseTestFiles}, wantFile: fuseTestFiles[2], wantRow: 2545, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, gotRow, err := itemLocation(tt.args.itemID, tt.args.files, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("itemLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFile != tt.wantFile {
				t.Errorf("itemLocation() gotFile = %v, want %v", gotFile, tt.wantFile)
			}
			if gotRow != tt.wantRow {
				t.Errorf("itemLocation() gotRow = %v, want %v", gotRow, tt.wantRow)
			}
		})
	}
}
