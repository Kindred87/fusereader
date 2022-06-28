package fusereader

import (
	"fmt"
	"path/filepath"
	"time"
)

const (
	parseBufferReceiveTimeout = time.Millisecond * 15000
	retrieveBufferSendTimeout = time.Millisecond * 2000
)

func parseWorker(file string, locate []FieldLocation, retrieve []FieldRetrieval, parseBuffer chan [][]string, retrieveBuffer chan Field) error {
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

func parseMatch(filename string, fields [][]string, find []FieldLocation) (bool, error) {
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

	for _, row := range fields {
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

func parseRetrieve(filename string, fields [][]string, retrieve []FieldRetrieval, buffer chan Field) error {
	fieldToSend := Field{}

	index, err := headerIndex(filename, headerItemID, []string{headerOperation}, 1)
	if err != nil {
		return fmt.Errorf("error while getting index for header %s in %s: %w", headerItemID, filepath.Base(filename), err)
	}

	if len(fields[0]) < index {
		return fmt.Errorf("length of first row for item in %s is less than the index of the header %s", filepath.Base(filename), headerItemID)
	}

	fieldToSend.ItemID = fields[0][index]

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

	for _, row := range fields {
		for specID, specIndex := range specIndices {
			if len(row) <= indexCache[specID] {
				continue
			}

			if retrieve[specIndex].Field.Matches(row[indexCache[specID]]) {
				retrieve[specIndex].Field.matchCount++

				if retrieve[specIndex].Field.matchCount >= int(retrieve[specIndex].Field.OnMatch) {
					fieldToSend.Header = retrieve[specIndex].Header.Key
					fieldToSend.SpecID = retrieve[specIndex].ID

					for _, offset := range retrieve[specIndex].FieldOffsets {
						fieldToSend.Value = row[indexCache[specID]+offset]

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
