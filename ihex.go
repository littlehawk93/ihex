// Package ihex provides a library of structs for reading and writing Intel HEX files
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
	maxRecordLen     = 521
	recordStartCode  = 58

	// HexFileTypeI8HEX the I8HEX file specification
	HexFileTypeI8HEX = 8,

	// HexFileTypeI16HEX the I16HEX file specification
	HexFileTypeI16HEX = 16,

	// HexFileTypeI16HEX the I32HEX file specification
	HexFileTypeI32HEX = 32,

	// RecordTypeData type of record that contains data.
	RecordTypeData = 0

	// RecordTypeEndOfFile type of record that indicates the end of the Intel HEX file.
	RecordTypeEndOfFile = 1

	// RecordTypeExtendedSegmentAddress type of record that contains a 16 bit segment base address. 
	// Useable by I16HEX files only. 
	// Multiply this record value by 16 and add to the address of all subsequent records.
	RecordTypeExtendedSegmentAddress = 2

	// RecordTypeStartSegmentAddress type of record that Initializes content of CS:IP registers for 80x86 processors.
	// Useable by I16HEX files only. 
	RecordTypeStartSegmentAddress = 3

	// RecordTypeExtendedLinearAddress type of record that contains a 16 bit segment base address. 
	// Useable by I32HEX files only. 
	// Use as the higher 16 bits of subsequent record addresses to form 32 bit addresses.
	RecordTypeExtendedLinearAddress = 4

	// RecordTypeStartLinearAddress type of record that stores the 32 bit value of the EIP register of the 80386 and higher Intel CPU.
	// Useable by I32HEX files only.
	RecordTypeStartLinearAddress = 5
)

// Record represents a single data record in an Intel HEX file.
type Record struct {
	RecordType    byte
	AddressOffset uint16
	Data          []byte
}

func (me *Record) write(writer io.Writer) error {

	buff := bytes.NewBufferString(":")

	err := me.writeNoErr(buff, fmt.Sprintf("%02X", len(me.Data)), nil)

	err = me.writeNoErr(buff, fmt.Sprintf("%04X", me.AddressOffset), err)

	err = me.writeNoErr(buff, fmt.Sprintf("%02X", me.RecordType), err)

	hexBytes := make([]byte, len(me.Data)*2)
	hex.Encode(hexBytes, me.Data)

	err = me.writeNoErr(buff, strings.ToUpper(string(hexBytes)), err)

	err = me.writeNoErr(buff, fmt.Sprintf("%02X", generateChecksum(me.Data, me.RecordType)), err)

	err = me.writeNoErr(buff, "\n", err)

	if err != nil {
		err = writer.Write(buff.Bytes())
	}

	return err
}

func (me *Record) writeNoErr(buff *bytes.Buffer, str string, err error) error {

	if err != nil {
		return err
	}

	_, err = buff.WriteString(str)

	return err
}

func (me *Record) read(scanner *bufio.Scanner) error {

	bytes := scanner.Bytes()

	if len(bytes) > maxRecordLen {
		return fmt.Errorf("Record size (%d) than maximum record size (%d)", len(bytes), maxRecordLen)
	}

	if bytes[0] != recordStartCode {
		return errors.New("Invalid start code for record")
	}

	byteCount, err := strconv.ParseInt(string(bytes[1:3]), 16, 8)

	if err != nil {
		return fmt.Errorf("Error parsing Byte Count: %s", err.Error())
	}

	actualByteCount := (len(bytes) - 11) / 2

	if byteCount > 255 {
		return errors.New("Invalid Byte Count provided")
	} else if int(byteCount) != actualByteCount {
		return fmt.Errorf("Record Byte Count (%d) does not match actual byte count (%d)", byteCount, actualByteCount)
	}

	me.Data = make([]byte, int(byteCount))

	address, err := strconv.ParseUint(string(bytes[3:7]), 16, 16)

	if err != nil {
		return fmt.Errorf("Error parsing Address: %s", err.Error())
	}

	me.AddressOffset = uint16(address)

	recordType, err := strconv.ParseUint(string(bytes[7:9]), 16, 8)

	if err != nil {
		return fmt.Errorf("Error parsing Record Type: %s", err.Error())
	}

	me.RecordType = byte(recordType)

	decodedBytes, err := hex.Decode(me.Data, bytes[9:bytesLen-2])

	if err != nil {
		return fmt.Errorf("Unable to parse record data: %s", err.Error())
	}

	if decodedBytes != len(me.Data) {
		return errors.New("Unable to parse record data: Incorrect number of bytes parsed")
	}

	recordChecksum, err := strconv.ParseUint(string(bytes[bytesLen-2:bytesLen]), 16, 8)

	if err != nil {
		return fmt.Errorf("Error parsing Checksum: %s", err.Error())
	}

	actualChecksum := generateChecksum(me.Data, me.RecordType)

	if byte(recordChecksum) != actualChecksum {
		return errors.New("Record checksum doesn't match computed checksum")
	}

	return nil
}

// I8HEX the I8HEX subfile type of the Intel HEX specification.
type HexFile struct {
	Type          byte
	records       []Record
	recordIndex   int
}

// Next advances the internal cursor to the next HEX record.
// Must be called at least once before calling the Record() method to initialize the internal cursor. 
// Returns false when there are no more records to read.
// Otherwise, returns true.
func (me *HexFile) Next() *Record {

	return ++me.recordIndex < len(me.records)
}

// Record retrieves the HEX record at the current position of the internal cursor
// Returns nil if the internal cursor hasn't been initialized or has moved past the last record.
func (me *HexFile) Record() *Record {

	if me.recordIndex < 0 || me.recordIndex >= len(me.records) {
		return nil
	}

	return me.records[me.recordIndex]
}

// ReadFrom reads HEX formatted record data from a reader.
// Returns any errors generated.
func (me *HexFile) ReadFrom(reader io.Reader) error {

	scanner := bufio.NewScanner(reader)

	records := make([]Record, 0)

	lineNumber := 1

	for scanner.Next() {

		var record Record

		err := record.read(scanner)

		if err != nil {
			if err == io.EOF {
				break
			} else {
				return make([]Record, 0), fmt.Errorf("Error on line %d: '%s'", lineNumber, err.Error())
			}
		}

		if len(records) > 0 && records[len(records) - 1].RecordType == RecordTypeEndOfFile {
			return make([]Record, 0), fmt.Errorf("Error on line %d: 'Record parsed after EOF record'", lineNumber)
		}

		if !validateHEXRecord(record, me.Type) {
			return make([]Record, 0), fmt.Errorf("Error on line %d: 'Invalid Record Type (%d)'", lineNumber, record.RecordType)
		}

		records = append(records, record)

		lineNumber++
	}

	if len(records) == 0 || records[len(records) - 1].RecordType != RecordTypeEndOfFile {
		return make([]Record, 0), fmt.Errorf("Error: No EOF record type parsed before EOF")
	}

	if err != nil {
		return err
	}

	me.records = records
	me.Reset()

	return nil
}

// Write writes this HEX file data to a writer.
// Returns any errors generated.
func (me *HexFile) WriteTo(writer io.Writer) error {

	for _, record := range me.records {

		err := record.write(writer)

		if err != nil {
			return err
		}
	}

	return nil
}

// AddRecord adds a new HEX record to this HEX file.
func (me *HexFile) AddRecord(record *Record) error {

	if !validateHEXRecord(record, me.Type) {
		return errors.New("Invalid record type for this HEX type")
	}

	if len(me.records) > 0 && me.records[len(me.records) - 1].RecordType == RecordTypeEndOfFile {
		return errors.New("Cannot add records after EOF record has been added")
	}

	me.records = append(me.records, *record)

	return nil
}

// Reset moves the internal cursor to begin reading records from the beginning.
func (me *HexFile) Reset() {

	me.recordIndex = -1
}

// NewHexFile creates a new empty HEX file.
func NewHexFile() *HexFile {

	var hex HexFile

	hex.records = make([]Record, 0)
	hex.Reset()

	return &hex
}

func validateHEXRecord(record *Record, hexType byte) bool {

	if record == nil {
		return false
	}

	isDataOrEOF := record.RecordType == RecordTypeData || record.RecordType == RecordTypeEndOfFile

	if hexType == HexFileTypeI8HEX {
		return isDataOrEOF
	} else if hexType == HexFileTypeI16HEX {
		return isDataOrEOF || record.RecordType == RecordTypeExtendedSegmentAddress || record.RecordType == RecordTypeExtendedSegmentAddress
	} else if hexType == HexFileTypeI32HEX {
		return isDataOrEOF || record.RecordType == RecordTypeExtendedLinearAddress || record.RecordType == RecordTypeStartLinearAddress
	} else {
		return false
	}
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
