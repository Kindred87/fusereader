package fusereader

// Field represents a field within a FUSE file.
type Field interface {
	SetSpecID(string)  // SetSpecID sets the ID of the field specification responsible for the retrieval of this field.
	SpecID() string    // SpecID returns the ID of the field specification responsible for the retrieval of this field.
	SetItemID(string)  // SetItemID sets the item ID associated with the field.
	ItemID() string    // ItemID returns the item ID associated with the field.
	SetHeader(string)  // SetHeader sets the column header for the field.
	Header() string    // Header returns the column header for the field.
	SetValue(string)   // SetValue sets the contents of the field.
	Value() string     // Value returns the contents of the field.
	SetFile(string)    // SetFile sets the filename of the spreadsheet the field was retrieved from.
	File() string      // File returns the filename of the spreadsheet the field was retrieved from.
	SetAddress(string) // SetAddress sets the address of the cell in A1 format.
	Address() string   // Address returns the address of the cell in A1 format.
}

// field represents a field within a FUSE file.
type field struct {
	specID  string // specID is the ID of the field specification responsible for the retrieval of this field.
	itemID  string // itemID is the item ID associated with the field.
	header  string // header is the column header for the field.
	value   string // value is the contents of the field.
	file    string // file is the filename of the spreadsheet the field was retrieved from.
	address string // address is the address of the cell in A1 format.
}

// SetSpecID sets the ID of the field specification responsible for the retrieval of this field.
func (f *field) SetSpecID(s string) {
	f.specID = s
}

// SpecID returns the ID of the field specification responsible for the retrieval of this field.
func (f field) SpecID() string {
	return f.specID
}

// SetItemID sets the item ID associated with the field.
func (f *field) SetItemID(s string) {
	f.itemID = s
}

// ItemID returns the item ID associated with the field.
func (f field) ItemID() string {
	return f.itemID
}

// SetHeader sets the column header for the field.
func (f *field) SetHeader(s string) {
	f.header = s
}

// Header returns the column header for the field.
func (f field) Header() string {
	return f.header
}

// SetValue sets the contents of the field.
func (f *field) SetValue(s string) {
	f.value = s
}

// Value returns the contents of the field.
func (f field) Value() string {
	return f.value
}

// SetFile sets the filename of the spreadsheet the field was retrieved from.
func (f *field) SetFile(s string) {
	f.file = s
}

// File returns the filename of the spreadsheet the field was retrieved from.
func (f field) File() string {
	return f.file
}

// SetAddress sets the address of the cell in A1 format.
func (f *field) SetAddress(s string) {
	f.address = s
}

// Address returns the address of the cell in A1 format.
func (f field) Address() string {
	return f.address
}
