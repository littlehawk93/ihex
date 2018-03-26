package ihex

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	i8HEXMaxAddress  = 65535
	i16HEXMaxAddress = 1048575
	i32HEXMaxAddress = 4294967295
	maxRecordLen     = 521
	recordStartCode  = 58

	// RecordTypeData - Record containing data
	RecordTypeData = 0

	// RecordTypeEndOfFile - Record indicating the end of the Intel HEX file
	RecordTypeEndOfFile = 1

	// RecordTypeExtendedSegmentAddress - Useable by I16HEX files only. Record containing a 16 bit segment base address. Multiple this record value by 16 and add to the address of all subsequent records
	RecordTypeExtendedSegmentAddress = 2

	// RecordTypeStartSegmentAddress - Useable by I16HEX files only. Initializes content of CS:IP registers for 80x86 processors
	RecordTypeStartSegmentAddress = 3

	// RecordTypeExtendedLinearAddress - Useable by I32HEX files only. Record containing a 16 bit segment base address. Use as the higher 16 bits of subsequent record addresses to form 32 bit addresses
	RecordTypeExtendedLinearAddress = 4

	// RecordTypeStartLinearAddress - Useable by I32HEX files only. Stores the 32 bit value of the EIP register of the 80386 and higher Intel CPU
	RecordTypeStartLinearAddress = 5
)

// File - Represents any of the file types of the Intel HEX file specification
type File interface {
	WriteFile(*os.File) error
	NextRecord() *Record
	AddRecord(*Record)
	Size() int
	Reset()
}

type recordOffsetIndex struct {
	AddressOffset uint32
	RecordIndex   int
}

// Record - Represents a single data record in an Intel HEX file
type Record struct {
	RecordType    byte
	AddressOffset uint16
	Data          []byte
}

func (me *Record) write(writer io.Writer) error {

	buff := bytes.NewBufferString(":")

	_, err := buff.WriteString(fmt.Sprintf("%02X", len(me.Data)))

	if err != nil {
		return err
	}

	_, err = buff.WriteString(fmt.Sprintf("%04X", me.AddressOffset))

	if err != nil {
		return err
	}

	_, err = buff.WriteString(fmt.Sprintf("%02X", me.RecordType))

	if err != nil {
		return err
	}

	hexBytes := make([]byte, len(me.Data)*2)

	hex.Encode(hexBytes, me.Data)

	_, err = buff.WriteString(strings.ToUpper(string(hexBytes)))

	if err != nil {
		return err
	}

	_, err = buff.WriteString(fmt.Sprintf("%02X", generateChecksum(me.Data, me.RecordType)))

	if err != nil {
		return err
	}

	_, err = buff.WriteString("\n")

	_, err = writer.Write(buff.Bytes())

	return err
}

// I8HEX - The I8HEX subfile type of the Intel HEX specification
type I8HEX struct {
	records       []Record
	addressOffset uint16
	recordIndex   int
}

// NextRecord - Retrieve the next HEX record from the I8HEX file. Nil if last record has already been read
func (me *I8HEX) NextRecord() *Record {

	if me.recordIndex >= len(me.records) {

		return nil
	}

	record := me.records[me.recordIndex]

	me.recordIndex++

	return &record
}

// WriteFile - Write this I8HEX to a file
func (me *I8HEX) WriteFile(file *os.File) error {

	for _, record := range me.records {

		err := record.write(file)

		if err != nil {
			return err
		}
	}

	return nil
}

// AddRecord - Add a new HEX record to this I8HEX file
func (me *I8HEX) AddRecord(record *Record) {

	me.records = append(me.records, *record)
}

// Size - Get the total count of records in this I8HEX file
func (me *I8HEX) Size() int {

	return len(me.records)
}

// Reset - Begin reading records from this I8HEX file from the beginning record
func (me *I8HEX) Reset() {

	me.recordIndex = 0
}

// I16HEX - The I16HEX subfile type of the Intel HEX specification
type I16HEX struct {
	records       []Record
	addressOffset uint32
	recordIndex   int
	offsetIndexes []recordOffsetIndex
}

// I32HEX - The I32HEX subfile type of the Intel HEX specification
type I32HEX struct {
	records       []Record
	addressOffset uint32
	recordIndex   int
	offsetIndexes []recordOffsetIndex
}

// NewI8HEX - Create a new I8HEX file by parsing data from a reader
func NewI8HEX(reader io.Reader) (*I8HEX, error) {

	var file I8HEX

	file.addressOffset = 0
	file.recordIndex = 0
	file.records = make([]Record, 0)

	bReader := bufio.NewReader(reader)

	lineNumber := 1

	for true {

		record, err := readRecord(bReader)

		if err != nil {

			if err == io.EOF {
				break
			} else {
				return nil, fmt.Errorf("Error on line %d: '%s'", lineNumber, err.Error())
			}
		}

		if record.RecordType != RecordTypeData && record.RecordType != RecordTypeEndOfFile {
			return nil, fmt.Errorf("Error on line %d: 'Invalid Record Type (%d)'", lineNumber, record.RecordType)
		}

		file.records = append(file.records, *record)

		lineNumber++
	}

	return &file, nil
}

// NewI16HEX - Create a new I16HEX file by parsing data from a reader
func NewI16HEX(reader *io.Reader) (*I16HEX, error) {

	return nil, nil
}

// NewI32HEX - Create a new I32HEX file by parsing data from a reader
func NewI32HEX(reader *io.Reader) (*I32HEX, error) {

	return nil, nil
}

func readRecord(reader *bufio.Reader) (*Record, error) {

	var record Record

	bytes := make([]byte, 0)

	bytesLen := 0

	reading := true

	for reading {

		tmpBytes, isPrefix, err := reader.ReadLine()

		if err != nil {
			return nil, err
		}

		bytes = append(bytes, tmpBytes[:]...)

		bytesLen = len(bytes)

		reading = isPrefix

		if bytesLen > maxRecordLen {
			return nil, errors.New("Maximum Record length exceeded")
		}
	}

	if bytes[0] != recordStartCode {

		return nil, errors.New("Invalid start code for record")
	}

	byteCount, err := strconv.ParseInt(string(bytes[1:3]), 16, 8)

	if err != nil {
		return nil, fmt.Errorf("Error parsing Byte Count: %s", err.Error())
	}

	actualByteCount := (len(bytes) - 11) / 2

	if byteCount > 255 {
		return nil, errors.New("Invalid Byte Count provided")
	} else if int(byteCount) != actualByteCount {
		return nil, fmt.Errorf("Record Byte Count (%d) does not match actual byte count (%d)", byteCount, actualByteCount)
	}

	record.Data = make([]byte, int(byteCount))

	address, err := strconv.ParseUint(string(bytes[3:7]), 16, 16)

	if err != nil {
		return nil, fmt.Errorf("Error parsing Address: %s", err.Error())
	}

	record.AddressOffset = uint16(address)

	recordType, err := strconv.ParseUint(string(bytes[7:9]), 16, 8)

	if err != nil {
		return nil, fmt.Errorf("Error parsing Record Type: %s", err.Error())
	}

	if recordType != RecordTypeData && recordType != RecordTypeEndOfFile && recordType != RecordTypeExtendedLinearAddress && recordType != RecordTypeExtendedSegmentAddress && recordType != RecordTypeStartLinearAddress && recordType != RecordTypeStartSegmentAddress {
		return nil, fmt.Errorf("Invalid Record Type (%d)", recordType)
	}

	record.RecordType = byte(recordType)

	decodedBytes, err := hex.Decode(record.Data, bytes[9:bytesLen-2])

	if err != nil {
		return nil, fmt.Errorf("Unable to parse record data: %s", err.Error())
	}

	if decodedBytes != len(record.Data) {
		return nil, errors.New("Unable to parse record data: Incorrect number of bytes parsed")
	}

	recordChecksum, err := strconv.ParseUint(string(bytes[bytesLen-2:bytesLen]), 16, 8)

	if err != nil {
		return nil, fmt.Errorf("Error parsing Checksum: %s", err.Error())
	}

	actualChecksum := generateChecksum(record.Data, record.RecordType)

	if byte(recordChecksum) != actualChecksum {
		return nil, errors.New("Record checksum doesn't match computed checksum")
	}

	return &record, nil
}

func generateChecksum(data []byte, recordType byte) byte {

	if recordType == RecordTypeEndOfFile {
		return 0xFF
	}

	sum := uint32(0)

	for _, d := range data {
		sum += uint32(d)
	}

	checksum := byte(((^sum) & 0x000000FF)) + 1

	return checksum
}
