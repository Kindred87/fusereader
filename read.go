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
func readWorker(file string, parseIfMatches FieldSpecification, parseBuffer chan [][]string) error {
	defer close(parseBuffer)

	fi, err := getFile(file)
	if err != nil {
		return fmt.Errorf("error while getting file pointer for %s: %w", filepath.Base(file), err)
	}

	keyHeaderIndex, err := headerIndex(file, parseIfMatches.KeyHeader, parseIfMatches.InGroupContaining)
	if err != nil {
		return fmt.Errorf("error while determining key header index for FieldSpecification %s: %w", parseIfMatches.ID, err)
	}

	recordTypeIndex, err := headerIndex(file, headerRecordType, []string{headerOperation})
	if err != nil {
		return fmt.Errorf("error while determining index of header %s: %w", headerRecordType, err)
	}

	rows, err := fi.Rows(worksheetFSItem)
	if err != nil {
		return fmt.Errorf("error while getting row iterator for %s: %w", filepath.Base(file), err)
	}

	var emptyRows int = 0
	var currentRow int = 0
	var cells []string
	var itemCache [][]string
	var itemCacheRow int
	var parseItem bool = false

	for rows.Next() {

		currentRow++

		if emptyRows > emptyRowMax {
			break
		}

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

		if parseIfMatches.Match(cells[keyHeaderIndex]) {
			parseIfMatches.matches++
			if parseIfMatches.OnNMatches == uint(parseIfMatches.matches) {
				parseItem = true
			}
		}

		if cells[recordTypeIndex] == itemRecordType {
			if parseItem {
				parseItem = false

				select {
				case parseBuffer <- itemCache[:itemCacheRow+1]:
				case <-time.After(parseBufferSendTimeout):
					return fmt.Errorf("reader for %s timed out on row %d while waiting to send to parse buffer", filepath.Base(file), currentRow)
				}
			}

			itemCacheRow = 0
		} else {
			itemCacheRow++
		}

		if itemCacheRow >= len(itemCache) {
			itemCache = append(itemCache, cells)
		} else {
			itemCache[itemCacheRow] = cells
		}

	}
	return nil
}
