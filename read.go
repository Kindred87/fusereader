package fusereader

import (
	"fmt"
	"path/filepath"
	"time"
)

const (
	emptyRowMax            = 50
	parseBufferSendTimeout = time.Millisecond * 2000
)

// readWorker reads items in the given file, sending items containing values matching the given specification to the parse
// buffer.
func readWorker(file string, locate FieldLocation, parseBuffer chan parseTarget) error {
	defer close(parseBuffer)

	fi, err := getFile(file)
	if err != nil {
		return fmt.Errorf("error while getting file pointer for %s: %w", filepath.Base(file), err)
	}

	keyHeaderIndex, err := headerIndex(file, locate.Header.Key, locate.Header.OthersInGroup, locate.Header.OnMatch)
	if err != nil {
		return fmt.Errorf("error while determining key header index for FieldSpecification %s: %w", locate.ID, err)
	}

	recordTypeIndex, err := headerIndex(file, headerRecordType, []string{headerOperation}, 1)
	if err != nil {
		return fmt.Errorf("error while determining index of header %s: %w", headerRecordType, err)
	}

	rows, err := fi.Rows(worksheetFSItem)
	if err != nil {
		return fmt.Errorf("error while getting row iterator for %s: %w", filepath.Base(file), err)
	}

	var currentRow int = 0
	var cells []string
	var emptyRows int = 0
	var parseItem bool = false
	var itemCacheRow int
	var itemCache [][]string

	for rows.Next() {
		currentRow++

		cells, err = rows.Columns()
		if err != nil {
			return fmt.Errorf("error while reading row %d in %s: %w", currentRow, filepath.Base(file), err)
		}

		if keyHeaderIndex > len(cells) || recordTypeIndex > len(cells) {
			emptyRows++
			continue
		} else if cells[keyHeaderIndex] == "" {
			emptyRows++
		} else {
			emptyRows = 0
		}

		if emptyRows > emptyRowMax {
			break
		}

		if cells[recordTypeIndex] == itemRecordType {
			if parseItem {
				parseItem = false

				t := parseTarget{}
				t.file = file
				t.beginningRow = currentRow - len(itemCache[:itemCacheRow+1])
				t.rowContents = append(t.rowContents, itemCache[:itemCacheRow+1]...)

				select {
				case parseBuffer <- t:
				case <-time.After(parseBufferSendTimeout):
					return fmt.Errorf("reader for %s timed out on row %d while waiting to send to parse buffer", filepath.Base(file), currentRow)
				}
			}

			itemCacheRow = 0
		} else {
			itemCacheRow++
		}

		if locate.Field.Matches(cells[keyHeaderIndex]) {
			locate.Field.matchCount++
			if locate.Field.matchCount >= locate.Field.OnMatch {
				parseItem = true
			}
		}

		if itemCacheRow >= len(itemCache) {
			itemCache = append(itemCache, cells)
		} else {
			itemCache[itemCacheRow] = cells
		}

	}
	return nil
}
