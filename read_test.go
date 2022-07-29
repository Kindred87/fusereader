package fusereader

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func Test_readWorker(t *testing.T) {
	cacheFile(fuseTestFiles[0])
	defer closeFiles()

	fi, err := getFile(fuseTestFiles[0])
	assert.Nil(t, err)

	err = buildHeaderCaches(fi)
	defer removeHeaderCaches()
	assert.Nil(t, err)

	validSpec := FieldLocation{
		ID: "Valid",
		Header: HeaderSpecification{
			Key:           headerItemID,
			OthersInGroup: []string{headerItemType},
		},
		Field: FieldSpecification{
			Matches: func(s string) bool { return s == "00046015128797" },
			OnMatch: 1,
		},
	}

	type args struct {
		file           string
		parseIfMatches FieldLocation
		parseBuffer    chan parseTarget
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
		c := make(chan parseTarget)

		tt.args.parseBuffer = c
		if tt.consumeBuffer {
			go consumeBuffer(c)
		} else if tt.checkBuffer {
			eg.Go(func() error { return checkBufferRowContents(c, tt.args.file, tt.args.parseIfMatches) })
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
func consumeBuffer(c chan parseTarget, checkFor ...FieldSpecification) {
	for {
		_, ok := <-c

		if !ok {
			break
		}
	}
}

// checkBufferRowContents returns a non-nil error if the channel receives items that do not match
// the given field specification.
func checkBufferRowContents(c chan parseTarget, file string, checkFor FieldLocation) error {
	keyHeaderIndex, err := headerIndex(file, checkFor.Header.Key, checkFor.Header.OthersInGroup, checkFor.Header.OnMatch)
	if err != nil {
		return fmt.Errorf("error while getting header index for header %s in %s: %w", checkFor.Header.Key, file, err)
	}

	foundNeedle := false

	for {
		v, ok := <-c

		if !ok {
			break
		}

		for _, row := range v.rowContents {
			if keyHeaderIndex < len(row) && checkFor.Field.Matches(row[keyHeaderIndex]) {
				checkFor.Field.matchCount++

				if checkFor.Field.matchCount == int(checkFor.Field.OnMatch) {
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

func Test_readworker_for_beginning_row(t *testing.T) {
	cacheFile(fuseTestFiles[0])
	defer closeFiles()

	fi, err := getFile(fuseTestFiles[0])
	assert.Nil(t, err)

	err = buildHeaderCaches(fi)
	defer removeHeaderCaches()
	assert.Nil(t, err)

	fieldLocations := []FieldLocation{
		{ID: "00011110603081", Header: HeaderSpecification{Key: "Item ID", OthersInGroup: []string{"Item Type"}}, Field: FieldSpecification{Matches: func(s string) bool { return s == "00011110603081" }}},
		{ID: "10011110603088", Header: HeaderSpecification{Key: "Item ID", OthersInGroup: []string{"Item Type"}}, Field: FieldSpecification{Matches: func(s string) bool { return s == "10011110603088" }}},
		{ID: "00077661003169", Header: HeaderSpecification{Key: "Item ID", OthersInGroup: []string{"Item Type"}}, Field: FieldSpecification{Matches: func(s string) bool { return s == "00077661003169" }}},
	}

	tests := []struct {
		name                 string
		fieldLocation        FieldLocation
		expectedBeginningRow int
	}{
		{name: "00011110603081", fieldLocation: fieldLocations[0], expectedBeginningRow: 4},
		{name: "10011110603088", fieldLocation: fieldLocations[1], expectedBeginningRow: 23},
		{name: "00077661003169", fieldLocation: fieldLocations[2], expectedBeginningRow: 8595},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := make(chan parseTarget)
			var eg errgroup.Group

			eg.Go(func() error { return readWorker(fuseTestFiles[0], tt.fieldLocation, c) })
			eg.Go(func() error { return checkBufferBeginningRow(c, tt.expectedBeginningRow) })

			err := eg.Wait()
			assert.Nil(t, err)
		})
	}
}

func checkBufferBeginningRow(c chan parseTarget, checkFor int) error {
	foundNeedle := false

	for {
		v, ok := <-c

		if !ok {
			break
		}

		if v.beginningRow == checkFor {
			foundNeedle = true
		}
	}

	if foundNeedle {
		return nil
	}

	return fmt.Errorf("did not locate %d", checkFor)
}
