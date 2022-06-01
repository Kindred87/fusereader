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
)

var (
	headerIndexCache map[string]map[string]int
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
