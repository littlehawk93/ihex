package ihex

// I16HEXFile is a HEX file in I16HEX format
type I16HEXFile struct {
	recordIndex int
	records     []Record
}

// GetType returns the file type of this HEX file
func (me *I16HEXFile) GetType() FileType {
	return I16HEX
}

// ReadNext advances to the next record in this HEX file and returns it with a boolean flag of true.
// If there are no more records in the file, a dummy record is returned along with the boolean flag of false.
func (me *I16HEXFile) ReadNext() (Record, bool) {

	if me.recordIndex+1 >= len(me.records) {
		return Record{}, false
	}

	me.recordIndex++
	return me.records[me.recordIndex], true
}

// Reset resets this file back to the first record in the file to be ready to read again.
func (me *I16HEXFile) Reset() {
	me.recordIndex = -1
}

// Add adds a new record to the end of this HEX file
// Returns an error if the record is incompatible with this file type
func (me *I16HEXFile) Add(r Record) error {

	if !r.validate(me.GetType()) {
		return &InvalidRecordTypeError{
			InvaildFileType:   me.GetType(),
			InvalidRecordType: r.Type,
		}
	}

	me.records = append(me.records, r)
	return nil
}

// AddRecords adds a set of records to the end of this HEX file
// Returns an error if any of the records are incompatible with this file type
func (me *I16HEXFile) AddRecords(r ...Record) error {
	for _, record := range r {
		if err := me.Add(record); err != nil {
			return err
		}
	}
	return nil
}

// NewI16HEXFile creates and initializes a new I16HEX file
// Returns the newly created I16HEX file
func NewI16HEXFile() *I16HEXFile {
	return &I16HEXFile{
		recordIndex: -1,
		records:     make([]Record, 0),
	}
}
