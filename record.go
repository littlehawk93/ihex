package ihex

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

const (
	// recordMaximumDataSize the largest size of the data payload of a record (in bytes)
	recordMaximumDataSize = 255

	// recordMaximumSizeChars the largest size of a record (including header data, checksum and starting character) when hexadecimal encoded
	recordMaximumSizeChars = 521

	// recordEOFChecksum is the default checksum value for an EOF record
	recordEOFChecksum = 0xFF

	// recordStartChar the starting character of a HEX record
	recordStartChar = ':'

	// Each of these defines where in the byte stream for a record to pull that record data property
	recordByteCountIndex   = 0
	recordAddressByteIndex = 1
	recordRecordTypeIndex  = 3
	recordDataIndex        = 4

	// recordHeaderAndChecksumSize defines how many bytes everythign in the record except the starting char and the data itself takes up
	recordHeaderAndChecksumSize = 5
)

// Record is a single record in an Intel HEX file.
// An IHEX record contains the following:
// The record type.
// A 16 bit record address (or address offset for I16HEX and I32HEX files).
// Up to 255 bytes of data.
// A checksum for validating record data integrity.
// This library handles checksum validation and generation automatically. Thus the checksum is excluded from the struct.
type Record struct {
	Type          RecordType
	AddressOffset uint16
	Data          []byte
}

// validate checks if this record belongs in the specified IHEX file format and that the length is correct.
// This always returns true if the record's type is RecordEOF or RecordData.
// If the file type is I16HEX, then this will return true if the record's type is RecordExtSegment or RecordStartSegment.
// If the file type is I32HEX, then this will return true if the record's type is RecordExtLinear or RecordStartLinear.
// Otherwise, returns false.
func (me Record) validate(fileType FileType) bool {
	return me.Type == RecordData || me.Type == RecordEOF || ((me.Type == RecordExtSegment || me.Type == RecordStartSegment) && fileType == I16HEX) || ((me.Type == RecordExtLinear || me.Type == RecordStartLinear) && fileType == I32HEX)
}

// write writes this record's data to a writer in valid IHEX format.
// Returns number of bytes written and any errors created during the writing process.
func (me Record) write(w io.Writer) (int, error) {

	buf := bytes.NewBufferString(fmt.Sprintf("%c%02X%04X%02X", recordStartChar, len(me.Data), me.AddressOffset, me.Type))

	hexBytes := make([]byte, len(me.Data)*2)
	hex.Encode(hexBytes, me.Data)

	if _, err := buf.WriteString(fmt.Sprintf("%s%02X\n", strings.ToUpper(string(hexBytes)), me.getChecksum())); err != nil {
		return 0, err
	}

	return w.Write(buf.Bytes())
}

// getChecksum generates the 8 bit checksum for this record.
// The IHEX specificaiton of the record checksum is that it is: "the two's complement of the least significant byte (LSB) of the sum of all decoded byte values in the record preceding the checksum".
// Returns the 1 byte (8 bit) checksum using the IHEX checksum specification.
func (me Record) getChecksum() byte {

	if me.Type == RecordEOF {
		return recordEOFChecksum
	}

	sum := uint32(0)

	for _, d := range me.Data {
		sum += uint32(d)
	}

	checksum := byte(((^sum) & 0x000000FF)) + 1

	return checksum
}
