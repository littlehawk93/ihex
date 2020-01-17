package ihex

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	writerMaximumI8HEXRecords  int64 = 65536
	writerMaximumI16HEXRecords int64 = 1048576
	writerMaximumI32HEXRecords int64 = 4294967296
)

// FileWriter writes a stream of bytes into HEX file format.
// The data is organized into records of fixed width with continuously incrementing addresses.
type FileWriter struct {
	recordSize  int
	recordCount int64
	buffer      []byte
	bufferIndex int
	fileType    FileType
	writer      io.Writer
	closed      bool
}

// Write writes the provided binary data in HEX format to the underlying writer.
// If len(p) exceeds the recordSize of this FileWriter, multiple records will be written to the stream
// Returns the number of bytes written (including the address and other header bytes of the record) or any errors encountered during writing.
func (me *FileWriter) Write(p []byte) (n int, err error) {

	if me.closed {
		return 0, errors.New("This FileWriter is closed")
	}

	sum := 0

	for i := 0; i < len(p); i++ {

		me.buffer[me.bufferIndex] = p[i]
		me.bufferIndex++

		if me.bufferIndex >= me.recordSize {
			me.bufferIndex = 0
			n, err = me.writeDataRecord(me.buffer)
			sum += n
			if err != nil {
				return sum, err
			}
		}
	}
	return sum, err
}

// Close closes this writer and flushes any remaining buffered data to the underling writer.
// This also closes the underlying writer if possible and writes the final EOF record to the writer.
// Returns any errors encountered during closing or when closing the underlying writer.
func (me *FileWriter) Close() error {

	if me.closed {
		return nil
	}

	me.closed = true

	if me.bufferIndex > 0 {
		for i := me.bufferIndex + 1; i < len(me.buffer); i++ {
			me.buffer[i] = 0
		}

		if _, err := me.writeDataRecord(me.buffer); err != nil {
			return err
		}
	}

	if c, ok := me.writer.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// NewFileWriter is equivalent to calling NewFileWriterType(w, recordSize, I32HEX)
// I32HEX is the default HEX format chosen to provide the largest supported record address range (4 GB) with maximum compatibility.
func NewFileWriter(w io.Writer, recordSize int) (*FileWriter, error) {

	return NewFileWriterType(w, recordSize, I32HEX)
}

// NewFileWriterType create and initialize a new FileWriter with the specified underlying writer to write HEX data into.
// All records written by this FileWriter will have data size of recordSize bytes.
// Returns a newly created and initialized FileWriter or an error if recordSize exceeds the maximum HEX data length (255 bytes)
func NewFileWriterType(w io.Writer, recordSize int, fileType FileType) (*FileWriter, error) {

	if recordSize > recordMaximumDataSize || recordSize <= 0 {
		return nil, fmt.Errorf("HEX record size cannot exceed %d bytes and must be greater than 0 bytes. Requested record size: %d bytes", recordMaximumDataSize, recordSize)
	}

	return &FileWriter{
		recordSize:  recordSize,
		recordCount: 0,
		buffer:      make([]byte, recordSize),
		bufferIndex: 0,
		fileType:    fileType,
		writer:      w,
		closed:      false,
	}, nil
}

// writeDataRecord handles writing a single record to the underlying writer.
// Automatically increments the FileWriter record count as new records are written.
// Automatically inserts address extension records as needed when record counts exceed I8HEX specifications.
// Returns the number of bytes written to the underlying writer and any errors that occurred during writing.
func (me *FileWriter) writeDataRecord(data []byte) (int, error) {

	if (me.fileType == I8HEX && me.recordCount >= writerMaximumI8HEXRecords) || (me.fileType == I16HEX && me.recordCount >= writerMaximumI16HEXRecords) || (me.fileType == I32HEX && me.recordCount >= writerMaximumI32HEXRecords) {
		return 0, fmt.Errorf("Maximum file record count for I%dHEX file exceeded", int(me.fileType))
	}

	address := uint16(me.recordCount & 0x00000000000000FF)
	sum := 0
	addressExtension := uint16(0)

	// evaluates for I32HEX and I16HEX. Whenever the maximum 16 bit address is reached
	// calculates the appropriate data value for the extended address record about to be written
	if address == 0 && me.recordCount > 0 {
		// Linear Segment Adddress (future data records' addresses get an additional upper 16 bits equal to this value to create a 32 bit address)
		if me.fileType == I32HEX {
			addressExtension = uint16((me.recordCount & 0x000000000000FF00) >> 16)
			// Extended Segment Address (future data records' addresses get offset by this value x 16)
		} else if me.fileType == I16HEX {
			addressExtension = uint16(me.recordCount / 16)
		}
	}

	// if an address extension was needed, this block of code generates the address extension record and writes it to the writer
	if addressExtension > 0 {

		b := make([]byte, 0)
		binary.BigEndian.PutUint16(b, addressExtension)

		// Set the appropriate extension record type depending on HEX file type
		t := RecordExtLinear
		if me.fileType == I16HEX {
			t = RecordExtSegment
		}

		r := Record{
			Type:          t,
			AddressOffset: 0,
			Data:          b,
		}

		n, err := r.write(me.writer)
		sum += n

		if err != nil {
			return sum, err
		}
	}

	me.recordCount++

	r := Record{
		Type:          RecordData,
		AddressOffset: address,
		Data:          data,
	}

	n, err := r.write(me.writer)
	return n + sum, err
}
