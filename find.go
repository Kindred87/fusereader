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
