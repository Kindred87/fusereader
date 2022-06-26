package fusereader

import (
	"fmt"
	"path/filepath"

	"github.com/emirpasic/gods/trees/avltree"
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
	headerCache          map[string]map[string][]int // headerCache contains the header index caches for one or more files.  If all files share the same header indices, then the key used will be the value of sharedHeaderCacheKey.
	headerGroupRootCache map[string]*avltree.Tree    // headerGroupRootCache stores the header group roots for each file within a binary tree.
)

const (
	sharedHeaderCacheKey = "shared" // sharedHeaderCacheKey is used as the file key for the header cache in situations where all files contain share the same header indices.
	headerRowMax         = 5        // headerRowMax describes the maximum number of rows by which the header row should have been found.
)

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

// buildHeaderCaches creates a cache of the headers in the given files used by header index calculation functions.
func buildHeaderCaches(files ...*excelize.File) error {
	if len(files) == 0 {
		return fmt.Errorf("no files were given")
	}

	headers, err := assembleHeaders(files)
	if err != nil {
		return err
	}

	headerCache = make(map[string]map[string][]int)

	if headersAreShared(headers) {
		headerCache[sharedHeaderCacheKey] = make(map[string][]int)

		for i, header := range headers[0] {
			headerCache[sharedHeaderCacheKey][header] = append(headerCache[sharedHeaderCacheKey][header], i)
		}

		if err = buildHeaderGroupRootCache(sharedHeaderCacheKey); err != nil {
			return fmt.Errorf("error while building header group root cache for shared headers: %w", err)
		}

	} else {
		for i, file := range files {
			headerCache[file.Path] = make(map[string][]int)

			for j, header := range headers[i] {
				headerCache[file.Path][header] = append(headerCache[file.Path][header], j)
			}

			if err = buildHeaderGroupRootCache(file.Path); err != nil {
				return fmt.Errorf("error while building header group root cache for %s: %w", filepath.Base(file.Path), err)
			}
		}
	}

	return nil
}

// assembleHeaders returns a slice containing the headers for all of the given files.
func assembleHeaders(files []*excelize.File) ([][]string, error) {
	headers := make([][]string, len(files))

	for i, file := range files {
		h, err := headersFrom(file)
		if err != nil {
			return nil, fmt.Errorf("error while getting headers from %s: %w", filepath.Base(file.Path), err)
		}

		headers[i] = h
	}

	return headers, nil
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

// headersAreShared returns true if all of the given headers are identical.
func headersAreShared(headers [][]string) bool {
	for _, h := range headers[1:] {
		if len(h) != len(headers[0]) {
			return false
		}
	}

	for _, h := range headers[1:] {
		for i, header := range h {
			if header != headers[0][i] {
				return false
			}
		}
	}

	return true
}

// buildHeaderGroupRootCache creates a cache of header group root indices for the given key for the main header
// index cache.
func buildHeaderGroupRootCache(cacheKey string) error {
	if _, exist := headerCache[cacheKey]; !exist {
		return fmt.Errorf("the key %s does not exist in the header cache", cacheKey)
	}

	if headerGroupRootCache == nil {
		headerGroupRootCache = make(map[string]*avltree.Tree)
	}

	tree := avltree.NewWithIntComparator()
	tree.Put(-1, -1) // Because the very first group starts at index 0.

	for _, index := range headerCache[cacheKey][headerNewGroupIndicator] {
		tree.Put(index, index)
	}

	headerGroupRootCache[cacheKey] = tree
	return nil
}

// removeHeaderCaches empties the header caches for garbage collection.
func removeHeaderCaches() {
	for k := range headerCache {
		headerCache[k] = nil
	}

	headerCache = nil

	for k := range headerGroupRootCache {
		headerGroupRootCache[k].Clear()
	}

	headerGroupRootCache = nil
}

// headerIndex returns the zero-based index of the given key header.
func headerIndex(file, keyHeader string, otherHeadersInGroup []string) (int, error) {
	if _, exist := headerCache[sharedHeaderCacheKey]; exist {
		file = sharedHeaderCacheKey
	}

	root, err := headerGroupRootIndex(file, append(otherHeadersInGroup, keyHeader))
	if err != nil {
		return 0, fmt.Errorf("error while getting index of group root for %s and %#v: %w", keyHeader, otherHeadersInGroup, err)
	}

	tree := avltree.NewWithIntComparator()

	for _, index := range headerCache[file][keyHeader] {
		tree.Put(index, index)
	}

	v, found := tree.Ceiling(root)
	tree.Clear()
	if found {
		return v.Key.(int), nil
	}

	return 0, fmt.Errorf("unable to determine index for %s in group containing %#v", keyHeader, otherHeadersInGroup)
}

// headerGroupRootIndex returns the zero-based index of the root of the group containing the given headers from the
// given file.
func headerGroupRootIndex(file string, headersInGroup []string) (int, error) {
	if headerCache == nil {
		return 0, fmt.Errorf("header cache is nil")
	} else if headerGroupRootCache == nil {
		return 0, fmt.Errorf("header group root cache is nil")
	}

	if _, exist := headerCache[sharedHeaderCacheKey]; exist {
		file = sharedHeaderCacheKey
	} else if _, exist := headerCache[file]; !exist {
		return 0, fmt.Errorf("file %s does not exist in the header cache", filepath.Base(file))
	}

	commonIndices := make(map[int]int)

	for i, header := range headersInGroup {
		indices, err := headerGroupRootIndices(file, header)
		if err != nil {
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

// headerGroupRootIndices returns the group root indices that the given header belongs to, within the header cache
// of the given key.
func headerGroupRootIndices(cacheKey string, header string) ([]int, error) {
	var indices []int

	for _, index := range headerCache[cacheKey][header] {
		node, found := headerGroupRootCache[cacheKey].Floor(index)
		if !found {
			return nil, fmt.Errorf("could not locate group root for %s at index %d", header, index)
		}

		indices = append(indices, node.Key.(int))
	}

	return indices, nil
}
