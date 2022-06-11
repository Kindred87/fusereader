package fusereader

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

func headerIndex(needle string, haystack []string) (int, bool) {
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

// headerGroupIndices returns a map of the given headers and the beginning index for the group/s they belong to.
func headerGroupIndices(headers []string) (map[string][]int, error) {
	indices := make(map[string][]int)

	groupRoot := 1

	for i, header := range headers {
		if header == headerNewGroupIndicator {
			groupRoot = i + 1
		}

		indices[header] = append(indices[header], groupRoot)
	}

	return indices, nil
}
