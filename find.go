package fusereader

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

const (
	worksheetFSItem = "FS_Item"
)

func itemLocation(itemID string, files []string, opts ...Option) (file string, row int, err error) {
	ctx := context.Background()
	readerCtx, cancelReaders := context.WithCancel(ctx)
	readBuffer := make(chan readRow, readWorkerPoolSize())

	var eg errgroup.Group
	var itemRow readRow

	eg.Go(func() error { return readFrom(files, worksheetFSItem, readerCtx, readBuffer) })
	eg.Go(func() error {
		r, err := parseBufferFor(itemID, headerItemID, readBuffer, cancelReaders)
		if err != nil {
			return fmt.Errorf("error while parsing: %w", err)
		}
		itemRow = r
		return nil
	})

	if err := eg.Wait(); err != nil {
		return "", 0, fmt.Errorf("error while searching for %s: %w", itemID, err)
	}

	return itemRow.file, itemRow.row, nil
}

func GetFields(files []string, find, retrieve []FieldSpecification, c chan Field, opts ...Option) error {

	return nil
}

// FieldSpecification describes one or more fields within a single header group by their expected location and contents.
type FieldSpecification struct {
	ID                string            // ID is an optional, though recommended field that may be used to uniquely identify a FieldSpecification instance.
	KeyHeader         string            // KeyHeader stores the header under which field values will be searched with Match.
	InGroupContaining []string          // InGroupContaining contains the headers that share a header group with the key header in order to distinguish which group is being referred to.  This mitigates conflicts arising from header reuse across multiple groups, such as 'Width'.
	Match             func(string) bool // Match returns true if the given field value under the key header is considered to be a match.
	Headers           []string          // Headers contains the headers under which fields along the matched row within the header group will be captured.
	matches           int               // matches stores the number of times Match has returned true.
	OnNMatches        uint              // OnNMatches is an optional field that describes the number of times Match should return true before capturing a field.  It is set to 1 by default, with a value of 2 indicating capture on the second match, and so on.  If N > 1, a field will not be captured unless Match returns true N times.
}

// Field represents a field within a FUSE file.
type Field struct {
	SpecID string // SpecID is the ID of the field specification responsible for the retrieval of this field.
	ItemID string // ItemID is the item ID associated with the field.
	Header string // Header is the column header for the field.
	Value  string // Value is the contents of the field.
}
