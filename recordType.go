package ihex

// RecordType defines what the type of a single record in a HEX file.
type RecordType byte

const (
	// RecordData indicates the record contains data and a 16-bit starting address for the data.
	// The byte count specifies number of data bytes in the record.
	RecordData RecordType = 0x00

	// RecordEOF indicates that this record is the end of the HEX file. Must occur exactly once.
	// The data field is empty (thus byte count is 00) and the address field is typically 0000.
	RecordEOF RecordType = 0x01

	// RecordExtSegment is the extended segment address record type. Usable by I16HEX files only.
	// The data field contains a 16-bit segment base address (thus byte count is always 02) compatible with 80x86 real mode addressing.
	// The address field (typically 0000) is ignored.
	// The segment address from the most recent 02 record is multiplied by 16 and added to each subsequent data record address to form the physical starting address for the data.
	// This allows addressing up to one megabyte of address space.
	RecordExtSegment RecordType = 0x02

	// RecordStartSegment is the start segment address record type. Usable by I16HEX files only.
	// For 80x86 processors, specifies the initial content of the CS:IP registers (i.e., the starting execution address).
	// The address field is 0000, the byte count is always 04, the first two data bytes are the CS value, the latter two are the IP value.
	RecordStartSegment RecordType = 0x03

	// RecordExtLinear is the extended linear address record type. Usable by I32HEX files only.
	// Allows for 32 bit addressing (up to 4GiB). The record's address field is ignored (typically 0000) and its byte count is always 02.
	// The two data bytes (big endian) specify the upper 16 bits of the 32 bit absolute address for all subsequent type 00 records; these upper address bits apply until the next 04 record.
	// The absolute address for a type 00 record is formed by combining the upper 16 address bits of the most recent 04 record with the low 16 address bits of the 00 record.
	// If a type 00 record is not preceded by any type 04 records then its upper 16 address bits default to 0000.
	RecordExtLinear RecordType = 0x04

	// RecordStartLinear is the start linear address record type. Usable by I32HEX files only.
	// The address field is 0000 (not used) and the byte count is always 04.
	// The four data bytes represent a 32-bit address value (big-endian).
	// In the case of 80386 and higher CPUs, this address is loaded into the EIP register.
	RecordStartLinear RecordType = 0x05
)
