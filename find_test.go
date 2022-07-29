package fusereader

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestGetFieldsForRetrievedValue(t *testing.T) {
	files, err := cacheFiles(fuseTestFiles)
	defer closeFiles()
	assert.Nil(t, err)

	err = buildHeaderCaches(files...)
	defer removeHeaderCaches()
	assert.Nil(t, err)

	c := make(chan field, 10)

	var eg errgroup.Group

	eg.Go(func() error { return checkFieldBuffer(c, "FREE_FROM -- Free from", "AY9", fuseTestFiles[0]) })

	err = GetFields([]string{fuseTestFiles[0]}, []FieldLocation{validFieldLocation()}, []FieldRetrieval{validRetrieveSpec()}, c)
	assert.Nil(t, err)

	close(c)

	err = eg.Wait()
	assert.Nil(t, err)
}

func checkFieldBuffer(buf chan field, value string, address string, file string) error {
	for {
		v, ok := <-buf
		if !ok {
			break
		}

		if v.Value() != value {
			return fmt.Errorf("expected %s, got %s for spec %s", value, v.Value(), v.SpecID())
		} else if v.Address() != address {
			return fmt.Errorf("expected %s, got %s for spec %s", address, v.Address(), v.SpecID())
		} else if v.File() != file {
			return fmt.Errorf("expected %s, got %s for spec %s", file, v.File(), v.SpecID())
		}
	}

	return nil
}

func TestGetFields(t *testing.T) {
	type args struct {
		files      []string
		locate     []FieldLocation
		retrieve   []FieldRetrieval
		readBuffer chan field
		opts       []Option
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "No retrieve spec", args: args{files: []string{fuseTestFiles[0]}, locate: []FieldLocation{validFieldLocation()}}, wantErr: true},
		{name: "No locate spec", args: args{files: []string{fuseTestFiles[0]}, retrieve: []FieldRetrieval{validRetrieveSpec()}}, wantErr: true},
		{name: "Bad file", args: args{files: []string{"bad_file.xlsx"}, locate: []FieldLocation{validFieldLocation()}, retrieve: []FieldRetrieval{validRetrieveSpec()}}, wantErr: true},
		{name: "Negative offset", args: args{files: []string{fuseTestFiles[0]}, locate: []FieldLocation{validFieldLocation()}, retrieve: []FieldRetrieval{validRetrieveSpecOverrideOffsets([]int{-2000})}}, wantErr: true},
		{name: "Out of range offset", args: args{files: []string{fuseTestFiles[0]}, locate: []FieldLocation{validFieldLocation()}, retrieve: []FieldRetrieval{validRetrieveSpecOverrideOffsets([]int{20000})}}, wantErr: true},
		{name: "Invalid header", args: args{files: []string{fuseTestFiles[0]}, locate: []FieldLocation{validFindSpecKeyHeaderOverride("Foo header")}, retrieve: []FieldRetrieval{validRetrieveSpec()}}, wantErr: true},
	}
	for _, tt := range tests {
		c := make(chan field)
		tt.args.readBuffer = c
		go consumeRetrievalBuffer(c)

		t.Run(tt.name, func(t *testing.T) {
			if err := GetFields(tt.args.files, tt.args.locate, tt.args.retrieve, tt.args.readBuffer, tt.args.opts...); (err != nil) != tt.wantErr {
				t.Errorf("GetFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func consumeRetrievalBuffer(c chan field) {
	for {
		_, ok := <-c

		if !ok {
			break
		}
	}
}

func validFieldLocation() FieldLocation {
	return FieldLocation{
		ID: "Location spec 01",
		Header: HeaderSpecification{
			Key:           headerItemID,
			OthersInGroup: []string{headerItemID},
			OnMatch:       1,
		},
		Field: FieldSpecification{
			Matches: func(s string) bool { return s == "00011110603081" },
		},
	}
}

func validFindSpecKeyHeaderOverride(h string) FieldLocation {
	s := validFieldLocation()
	s.Header.Key = h

	return s
}

func validRetrieveSpec() FieldRetrieval {
	return FieldRetrieval{
		ID: "Retrieve spec 01",
		Header: HeaderSpecification{
			Key:           "Allergen Type Code",
			OthersInGroup: []string{"Level Of Containment"},
			OnMatch:       1,
		},
		Field: FieldSpecification{
			Matches: func(s string) bool { return strings.Contains(s, "Soybean") },
		},
		FieldOffsets: []int{1},
	}
}

func validRetrieveSpecOverrideOffsets(o []int) FieldRetrieval {
	f := validRetrieveSpec()
	f.FieldOffsets = o

	return f
}
