package ihex

import "fmt"

// InvalidRecordTypeError error indicating a record type is incompatible with the HEX file format the record was found in
type InvalidRecordTypeError struct {
	InvalidRecordType RecordType
	InvaildFileType   FileType
}

// Error returns the error message for this error
func (me *InvalidRecordTypeError) Error() string {
	return fmt.Sprintf("Record Type %02X is not valid for I%dHEX files", byte(me.InvalidRecordType), int(me.InvaildFileType))
}

// InvalidRecordError error indicating that a HEX record is formatted incorrectly in some way
type InvalidRecordError struct {
	Message string
}

// Error returns the error message for this error
func (me *InvalidRecordError) Error() string {
	return fmt.Sprintf("Record formatted incorrectly: %s", me.Message)
}

// IndexedRecordError an error that occurred at a particular record index
type IndexedRecordError struct {
	RecordError error
	Index       int
}

// Error returns the error message for this error
func (me *IndexedRecordError) Error() string {
	return fmt.Sprintf("Error occurred on record at index %d: %s", me.Index, me.RecordError.Error())
}
