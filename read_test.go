package fusereader

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func Test_readWorker(t *testing.T) {
	cacheFile(fuseTestFiles[0])

	fi, err := getFile(fuseTestFiles[0])
	assert.Nil(t, err)

	err = buildHeaderCaches(fi)
	assert.Nil(t, err)

	validSpec := FieldSpecification{
		ID:                "Valid",
		KeyHeader:         headerItemID,
		InGroupContaining: []string{headerItemType},
		Match:             func(s string) bool { return s == "00046015128797" },
		OnNMatches:        1,
	}

	type args struct {
		file           string
		parseIfMatches FieldSpecification
		parseBuffer    chan [][]string
	}
	tests := []struct {
		name          string
		args          args
		wantErr       bool
		consumeBuffer bool
		checkBuffer   bool
	}{
		{name: "Valid", args: args{file: fuseTestFiles[0], parseIfMatches: validSpec}, wantErr: false, consumeBuffer: false, checkBuffer: true},
	}

	var eg errgroup.Group

	for _, tt := range tests {
		c := make(chan [][]string)

		tt.args.parseBuffer = c
		if tt.consumeBuffer {
			go consumeBuffer(c)
		} else if tt.checkBuffer {
			eg.Go(func() error { return checkBuffer(c, tt.args.file, tt.args.parseIfMatches) })
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := readWorker(tt.args.file, tt.args.parseIfMatches, tt.args.parseBuffer); (err != nil) != tt.wantErr {
				t.Errorf("readWorker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if err := eg.Wait(); err != nil {
		t.Error(err)
	}
}

// consumeBuffer continuously empties the given buffer until it is closed.
func consumeBuffer(c chan [][]string, checkFor ...FieldSpecification) {
	for {
		_, ok := <-c

		if !ok {
			break
		}
	}
}

// checkBuffer returns a non-nil error if the channel receives items that do not match
// the given field specification.
func checkBuffer(c chan [][]string, file string, checkFor FieldSpecification) error {
	keyHeaderIndex, err := headerIndex(file, checkFor.KeyHeader, checkFor.InGroupContaining)
	if err != nil {
		return fmt.Errorf("error while getting header index for header %s in %s: %w", checkFor.KeyHeader, file, err)
	}

	foundNeedle := false

	for {
		v, ok := <-c

		if !ok {
			break
		}

		for _, row := range v {
			if keyHeaderIndex < len(row) && checkFor.Match(row[keyHeaderIndex]) {
				checkFor.matches++

				if checkFor.matches == int(checkFor.OnNMatches) {
					foundNeedle = true
				}
			}
		}
		if !foundNeedle {
			return fmt.Errorf("could not locate value specified by %s in %s", checkFor.ID, file)
		}
	}

	return nil
}
