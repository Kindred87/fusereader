package fusereader

// FieldLocation provides a specification for a field value that is used to identify an item of interest.
type FieldLocation struct {
	ID     string              // ID uniquely identifies a FieldSpecification instance.
	Header HeaderSpecification // Header contains the header specification.
	Field  FieldSpecification  // Field contains the field specification.
}

// HeaderSpecification provides a specification for a specific header within a given file.
type HeaderSpecification struct {
	Key           string   // Key is the key header of the specification, directly under which fields will be searched.
	OthersInGroup []string // OthersInGroup contains other headers in the same group as the key header.  These are used to distinguish between key headers contained in multiple different groups.
	OnMatch       int      // OnMatch describes which identified header should be referenced, if there are multiple identified.  A value less than or equal to one will result in the first identified header being referenced, a value of two the second identified header, and so on.  This value is most useful for key headers that are in multiple header groups containing the same sets of headers.
}

// FieldSpecification provides a specification for identifying a field of interest.
type FieldSpecification struct {
	Matches    func(string) bool // Matches returns true if the given field value under the key header is considered to be a match.
	matchCount int               // matchCount stores the number of times Matches has returned true.
	OnMatch    int               // OnMatch describes the number of times Match should return true before considering a match to be the field of interest.  A value less than or equal to 1 indicates the first match, a value of 2 the second match, and so on.  If N > 1, a field will not be captured unless Match returns true N times.
}

// FieldRetrieval is used to specify fields for retrieval.
type FieldRetrieval struct {
	ID           string              // ID uniquely identifies a FieldRetrieval instance.
	Header       HeaderSpecification // Header contains the header specification.
	Field        FieldSpecification  // Spec identifies a field from which offset fields will be retrieved.
	FieldOffsets []int               // RetrievalOffsets is a slice of right-facing offsets from the field described by Spec.  Fields at the offsets will be retrieved.
}

// NewFieldLocationAll returns a field location object containing the given fields and sub-fields.
//
// For more information on field location fields, please see FieldLocation, HeaderSpecification, and FieldSpecification.
func NewFieldLocationAll(id, keyHeader string, otherHeaders []string, headerOnMatch int, fieldMatches func(string) bool, fieldOnMatch int) FieldLocation {
	return FieldLocation{
		ID: id,
		Header: HeaderSpecification{
			Key:           keyHeader,
			OthersInGroup: otherHeaders,
			OnMatch:       headerOnMatch,
		},
		Field: FieldSpecification{
			Matches: fieldMatches,
			OnMatch: fieldOnMatch,
		},
	}
}

// NewFieldLocationShort returns a field location object containing the given fields.
//
// For more information on field location fields, please see FieldLocation, HeaderSpecification, and FieldSpecification.
func NewFieldLocationShort(id string, header HeaderSpecification, field FieldSpecification) FieldLocation {
	return NewFieldLocationAll(id, header.Key, header.OthersInGroup, header.OnMatch, field.Matches, field.OnMatch)
}

// NewHeaderSpecification returns a header specification object containing the given fields.
func NewHeaderSpecification(key string, othersInGroup []string, onMatch int) HeaderSpecification {
	return HeaderSpecification{
		Key:           key,
		OthersInGroup: othersInGroup,
		OnMatch:       onMatch,
	}
}

// NewFieldSpecification returns a field specification object containing the given fields.
func NewFieldSpecification(matches func(string) bool, onMatch int) FieldSpecification {
	return FieldSpecification{
		Matches: matches,
		OnMatch: onMatch,
	}
}

// NewFieldRetrievalAll returns a field retrieval object containing the given fields and sub-fields.
//
// For more information on field retrieval fields, please see FieldRetrieval, HeaderSpecification, and FieldSpecification.
func NewFieldRetrievalAll(id, keyHeader string, otherHeaders []string, headerOnMatch int, fieldMatches func(string) bool, fieldOnMatch int, fieldOffsets []int) FieldRetrieval {
	return FieldRetrieval{
		ID: id,
		Header: HeaderSpecification{
			Key:           keyHeader,
			OthersInGroup: otherHeaders,
			OnMatch:       headerOnMatch,
		},
		Field: FieldSpecification{
			Matches: fieldMatches,
			OnMatch: fieldOnMatch,
		},
		FieldOffsets: fieldOffsets,
	}
}

// NewFieldRetrievalShort returns a field retrieval object containing the given fields.
func NewFieldRetrievalShort(id string, header HeaderSpecification, field FieldSpecification, fieldOffsets []int) FieldRetrieval {
	return NewFieldRetrievalAll(id, header.Key, header.OthersInGroup, header.OnMatch, field.Matches, field.OnMatch, fieldOffsets)
}
