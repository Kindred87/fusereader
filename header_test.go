package fusereader

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
