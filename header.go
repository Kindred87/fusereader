package fusereader

import (
	"fmt"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

// header keys

const (
	headerRecordType              = "RECORD TYPE"
	headerOperation               = "OPERATION"
	headerImportItem              = "IMPORT ITEM?"
	headerInformationProviderGLN  = "Information Provider GLN"
	headerInformationProviderName = "Information Provider Name"
	headerItemType                = "Item Type"
	headerItemID                  = "Item ID"
	headerNewGroupIndicator       = "Indicator for New Group"
)

var (
	headerIndexCache map[string]map[string]int
)

const (
	headerRowMax = 5 // headerRowMax describes the maximum number of rows by which the header row should have been found.
)

func cacheHeaders(file string, headers []string) {
	if headerIndexCache == nil {
		headerIndexCache = make(map[string]map[string]int)
	}

	headerIndexCache[file] = make(map[string]int)

	for i, header := range headers {
		headerIndexCache[file][header] = i
	}
}

func indexForHeader(needle string, haystack []string) (int, bool) {
	for i, header := range haystack {
		if header == needle {
			return i, true
		}
	}

	return 0, false
}

func valueForHeader(header, file string, haystack []string) (string, bool) {
	index, exist := headerIndexCache[file][header]
	if !exist || index >= len(haystack) {
		return "", false
	}

	return haystack[index], true
}

// isHeaderRow returns true if the given slice contains the prefix expected in a FUSE header row.
func isHeaderRow(s []string) bool {
	if len(s) == 0 {
		return false
	}

	prefix := headerRowPrefix()

	if len(s) < len(prefix) {
		return false
	}

	for i, p := range prefix {
		if s[i] != p {
			return false
		}
	}

	return true
}

// headerRowPrefix returns the headers that a FUSE header row should begin with.
func headerRowPrefix() []string {
	var out []string

	out = append(out, headerRecordType)
	out = append(out, headerOperation)
	out = append(out, headerImportItem)
	out = append(out, headerInformationProviderGLN)
	out = append(out, headerInformationProviderName)
	out = append(out, headerItemType)
	out = append(out, headerItemID)

	return out
}

// headersFrom returns the contents of the header row in the given file.
func headersFrom(file *excelize.File) ([]string, error) {
	rows, err := file.Rows(worksheetFSItem)
	if err != nil {
		return nil, fmt.Errorf("error while initiating row iterator for %s: %w", filepath.Base(file.Path), err)
	}

	currRow := 0
	for rows.Next() {
		currRow++

		if currRow >= headerRowMax {
			return nil, fmt.Errorf("could not locate header row in %s", filepath.Base(file.Path))
		}

		r, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("error while reading row %d in %s: %w", currRow, filepath.Base(file.Path), err)
		}

		if isHeaderRow(r) {
			return r, nil
		}
	}

	return nil, fmt.Errorf("could not locate header row in %s", filepath.Base(file.Path))
}

// headerGroupIndices returns a map of the given headers and the beginning index for the group/s they belong to.
func headerGroupIndices(headers []string) (map[string][]int, error) {
	indices := make(map[string][]int)

	groupRoot := 1

	for i, header := range headers {
		if header == headerNewGroupIndicator {
			groupRoot = i
		}

		indices[header] = append(indices[header], groupRoot)
	}

	return indices, nil
}

// headerGroupIndex returns the zero-based index of the given key header.
func headerIndex(headerRow []string, keyHeader string, otherHeadersInGroup []string) (int, error) {
	groupIndices, err := headerGroupIndices(headerRow)
	if err != nil {
		return 0, fmt.Errorf("error while getting header group indices: %w", err)
	}

	root, err := headerGroupRootIndex(groupIndices, append(otherHeadersInGroup, keyHeader))
	if err != nil {
		return 0, fmt.Errorf("error while getting index of group root for %s and %#v: %w", keyHeader, otherHeadersInGroup, err)
	}

	for i := root; i < len(headerRow); i++ {
		if headerRow[i] == keyHeader {
			return i, nil
		}
	}

	return 0, fmt.Errorf("unable to determine index for %s in group containing %#v", keyHeader, otherHeadersInGroup)
}

// headerGroupRootIndex returns the zero-based index of the root of the group containing the key header
func headerGroupRootIndex(groupIndices map[string][]int, headersInGroup []string) (int, error) {
	commonIndices := make(map[int]int)

	for i, header := range headersInGroup {
		indices, exist := groupIndices[header]
		if !exist {
			return 0, fmt.Errorf("could not locate %s among the given headers", header)
		}

		matched := false
		if i == 0 {
			matched = true
		}

		for _, index := range indices {
			if _, exist := commonIndices[index]; exist {
				matched = true
			} else {
				commonIndices[index] = 0
			}

			commonIndices[index]++
		}

		if !matched {
			return 0, fmt.Errorf("could not find common group for %s and %#v", header, headersInGroup[:i])
		}
	}

	for k, v := range commonIndices {
		if v == len(headersInGroup) {
			return k, nil
		}
	}

	return 0, fmt.Errorf("unable to determine index for group containing %#v", headersInGroup)
}
