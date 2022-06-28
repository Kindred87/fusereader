package fusereader

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sync/errgroup"
)

const (
	worksheetFSItem = "FS_Item"
	itemRecordType  = "ITEM"
)

// FieldLocation provides a specification for a field value that is used to identify an item of interest.
type FieldLocation struct {
	ID     string              // ID uniquely identifies a FieldSpecification instance.
	Header HeaderSpecification // Header contains the header specification.
	Field  FieldSpecification  // Field contains the field specification.
}

// HeaderSpecification provides a specification for a specific header within a given file.
type HeaderSpecification struct {
	Key           string   // Key is the key header of the specification, directly under which fields will be searched.
	OthersInGroup []string // OthersInGroup contains other headers in the same group as the key header.  These are used to distinguish between key headers contained in multiple different groups.
	OnMatch       int      // OnMatch describes which identified header should be referenced, if there are multiple identified.  A value less than or equal to one will result in the first identified header being referenced, a value of two the second identified header, and so on.  This value is most useful for key headers that are in multiple header groups containing the same sets of headers.
}

// FieldSpecification provides a specification for identifying a field of interest.
type FieldSpecification struct {
	Matches    func(string) bool // Matches returns true if the given field value under the key header is considered to be a match.
	matchCount int               // matchCount stores the number of times Matches has returned true.
	OnMatch    uint              // OnMatch describes the number of times Match should return true before considering a match to be the field of interest.  A value less than or equal to 1 indicates the first match, a value of 2 the second match, and so on.  If N > 1, a field will not be captured unless Match returns true N times.
}

// FieldRetrieval is used to specify fields for retrieval.
type FieldRetrieval struct {
	ID           string              // ID uniquely identifies a FieldRetrieval instance.
	Header       HeaderSpecification // Header contains the header specification.
	Field        FieldSpecification  // Spec identifies a field from which offset fields will be retrieved.
	FieldOffsets []int               // RetrievalOffsets is a slice of right-facing offsets from the field described by Spec.  Fields at the offsets will be retrieved.
}

// Field represents a field within a FUSE file.
type Field struct {
	SpecID string // SpecID is the ID of the field specification responsible for the retrieval of this field.
	ItemID string // ItemID is the item ID associated with the field.
	Header string // Header is the column header for the field.
	Value  string // Value is the contents of the field.
}

func GetFields(files []string, locate []FieldLocation, retrieve []FieldRetrieval, readBuffer chan Field, opts ...Option) (err error) {
	if err := validateParametersForCaching(files, locate, retrieve, readBuffer); err != nil {
		return fmt.Errorf("error while validating parameters: %w", err)
	}
	defer func() {
		removeHeaderCaches()

		cErr := closeFiles()
		if err == nil && cErr != nil {
			err = fmt.Errorf("error while closing files: %w", cErr)
		}
	}()

	if err = buildCaches(files); err != nil {
		return fmt.Errorf("error while building caches: %w", err)
	}

	if err := validateParametersForSearching(files, locate, retrieve); err != nil {
		return fmt.Errorf("error while validating parameters: %w", err)
	}

	var eg errgroup.Group

	for _, file := range files {
		f := file
		c := make(chan [][]string, 2)
		eg.Go(func() error { return readWorker(f, locate[0], c) })
		eg.Go(func() error { return parseWorker(f, locate, retrieve, c, readBuffer) })
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("error while retrieving fields: %w", err)
	}

	return nil
}

// buildCaches builds the file and header caches using the given files.
func buildCaches(files []string) error {
	cachedFiles, err := cacheFiles(files)
	if err != nil {
		return fmt.Errorf("error while loading files: %w", err)
	}

	if err := buildHeaderCaches(cachedFiles...); err != nil {
		return fmt.Errorf("error while loading headers: %w", err)
	}

	return nil
}

// validateParametersForCaching returns a non-nil error if it detects a fatal error with the given parameters in regards
// to building file and header caches.
func validateParametersForCaching(files []string, locate []FieldLocation, retrieve []FieldRetrieval, readBuffer chan Field) error {
	if len(files) == 0 {
		return fmt.Errorf("no files were given")
	} else if len(locate) == 0 {
		return fmt.Errorf("locate is empty")
	} else if len(retrieve) == 0 {
		return fmt.Errorf("retrieve is empty")
	} else if readBuffer == nil {
		return fmt.Errorf("the read buffer is nil")
	}

	return nil
}

// validateParametersForSearching returns a non-nil error if it detects a fatal error with the given parameters in regards
// to performing a search.
func validateParametersForSearching(files []string, locate []FieldLocation, retrieve []FieldRetrieval) error {
	if err := validateFieldLocations(locate, files); err != nil {
		return err
	}

	if err := validateFieldRetrievals(retrieve, files); err != nil {
		return err
	}

	return nil
}

// validateFieldLocations returns a non-nil error if it detects a fatal error with the given field specs in regards
// to performing a search.
func validateFieldLocations(locate []FieldLocation, files []string) error {
	for _, l := range locate {
		for _, file := range files {
			_, err := headerIndex(file, l.Header.Key, l.Header.OthersInGroup, l.Header.OnMatch)
			if err != nil {
				return fmt.Errorf("error while getting index for header %s in %s: %w", l.Header.Key, filepath.Base(file), err)
			}
		}
	}

	return nil
}

// validateFieldRetrievals returns a non-nil error if it detects a fatal error with the given field retrievals in regards
// to performing a search.
func validateFieldRetrievals(retrieve []FieldRetrieval, files []string) error {
	headerCounts := make(map[string]int)

	for _, file := range files {
		c, err := headerCountIn(file)
		if err != nil {
			return fmt.Errorf("error while getting number of headers in %s: %w", filepath.Base(file), err)
		}

		headerCounts[file] = c
	}

	for _, r := range retrieve {
		for _, file := range files {

			index, err := headerIndex(file, r.Header.Key, r.Header.OthersInGroup, r.Header.OnMatch)
			if err != nil {
				return fmt.Errorf("error while getting index for header %s in %s: %w", r.Header.Key, filepath.Base(file), err)
			}

			for _, offset := range r.FieldOffsets {
				if index+offset < 0 {
					return fmt.Errorf("offset %d for field retrieval with spec ID %s results in a header index of %d", offset, r.ID, index+offset)
				} else if headerCounts[file] <= index+offset {
					return fmt.Errorf("offset %d for field retrieval with spec ID %s results in a header index of %d, exceeding the header count of %d in %s", offset, r.ID, index+offset, headerCounts[file], filepath.Base(file))
				}

			}
		}
	}

	return nil
}
