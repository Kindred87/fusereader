package fusereader

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
)

const (
	parseBufferReceiveTimeout = time.Millisecond * 15000
	retrieveBufferSendTimeout = time.Millisecond * 2000
)

// parseTarget contains item information for consumption by parsing functions.
type parseTarget struct {
	file         string     // file is the filename of the originating spreadsheet that rowContents was read from.
	beginningRow int        // beginningRow is the first row in the originating spreadsheet that rowContents was read from.
	rowContents  [][]string // rowContents are the rows for a particular item as read from the spreadsheet.
}

// parseWorker identifies matching items in the given parse buffer and sends retrieved fields to the given retrieval buffer.
//
// The given file should match the file being read by the function sending into the parse buffer.
func parseWorker(file string, locate []FieldLocation, retrieve []FieldRetrieval, parseBuffer chan parseTarget, retrieveBuffer chan field) error {
	for {
		select {
		case v, ok := <-parseBuffer:
			if !ok {
				return nil
			}

			matches, err := parseMatch(file, v, locate)
			if err != nil {
				return fmt.Errorf("error while checking for parse match in %s: %w", file, err)
			}

			if matches {
				if err := parseRetrieve(file, v, retrieve, retrieveBuffer); err != nil {
					return fmt.Errorf("error while parsing to retrieve values")
				}
			}

		case <-time.After(parseBufferReceiveTimeout):
			return fmt.Errorf("parse buffer timed out while waiting for receipt")
		}

	}
}

// parseMatch returns true if fields contains fields specified by the contents of find.
func parseMatch(filename string, target parseTarget, find []FieldLocation) (bool, error) {
	indexCache := make(map[string]int)
	specIndices := make(map[string]int)

	for i, spec := range find {
		index, err := headerIndex(filename, spec.Header.Key, spec.Header.OthersInGroup, spec.Header.OnMatch)
		if err != nil {
			return false, fmt.Errorf("error while getting index of key header for spec %s in file %s: %w", spec.ID, filepath.Base(filename), err)
		}

		indexCache[spec.ID] = index
		specIndices[spec.ID] = i
	}

	for _, row := range target.rowContents {
		for specID, index := range indexCache {
			if len(row) <= index {
				continue
			}

			if find[specIndices[specID]].Field.Matches(row[index]) {
				find[specIndices[specID]].Field.matchCount++

				if find[specIndices[specID]].Field.matchCount >= int(find[specIndices[specID]].Field.OnMatch) {
					delete(indexCache, specID)
					delete(specIndices, specID)
				}
			}
		}
	}

	return len(indexCache) == 0, nil
}

// parseRetrieve retrieves values specified by retrieve and sends them over the given buffer.
func parseRetrieve(filename string, target parseTarget, retrieve []FieldRetrieval, buffer chan field) error {
	fieldToSend := field{}

	index, err := headerIndex(filename, headerItemID, []string{headerOperation}, 1)
	if err != nil {
		return fmt.Errorf("error while getting index for header %s in %s: %w", headerItemID, filepath.Base(filename), err)
	}

	if len(target.rowContents[0]) < index {
		return fmt.Errorf("length of first row for item in %s is less than the index of the header %s", filepath.Base(filename), headerItemID)
	}

	fieldToSend.SetItemID(target.rowContents[0][index])

	specIndices := make(map[string]int)
	indexCache := make(map[string]int)

	for i, r := range retrieve {
		specIndices[r.ID] = i

		index, err := headerIndex(filename, r.Header.Key, r.Header.OthersInGroup, r.Header.OnMatch)
		if err != nil {
			return fmt.Errorf("error while getting index of key header for spec %s in file %s: %w", r.ID, filepath.Base(filename), err)
		}

		indexCache[r.ID] = index
	}

	for i, row := range target.rowContents {
		for specID, specIndex := range specIndices {
			if len(row) <= indexCache[specID] {
				continue
			}

			if retrieve[specIndex].Field.Matches(row[indexCache[specID]]) {
				retrieve[specIndex].Field.matchCount++

				if retrieve[specIndex].Field.matchCount >= int(retrieve[specIndex].Field.OnMatch) {
					fieldToSend.SetFile(target.file)

					fieldToSend.SetHeader(retrieve[specIndex].Header.Key)
					fieldToSend.SetSpecID(retrieve[specIndex].ID)

					for _, offset := range retrieve[specIndex].FieldOffsets {
						fieldToSend.SetValue(row[indexCache[specID]+offset])

						a, err := excelize.CoordinatesToCellName(indexCache[specID]+offset+1, target.beginningRow+i)
						if err != nil {
							return fmt.Errorf("error while converting column %d and row %d to a cell name: %w", indexCache[specID]+offset, target.beginningRow, err)
						}

						fieldToSend.SetAddress(a)

						select {
						case buffer <- fieldToSend:
						case <-time.After(retrieveBufferSendTimeout):
							return fmt.Errorf("timeout while waiting to send to retrieve buffer for spec ID %s in %s", retrieve[specIndex].ID, filepath.Base(filename))
						}
					}

				}
			}
		}
	}

	return nil
}
