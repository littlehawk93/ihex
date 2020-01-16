package ihex

// FileType defines which HEX format a HEX file is in.
// Its integer value corresponds to the HEX format.
type FileType byte

const (
	// I8HEX is the I8HEX file format.
	// The I8HEX file format supports up to 16 bit memory addresses.
	I8HEX FileType = 8

	// I16HEX is the I16HEX file format.
	// The I16HEX file format supports up to 20 bit memory addresses.
	I16HEX FileType = 16

	// I32HEX is the I32HEX file format.
	// The I32HEX file format supports up to 32 bit memory addresses.
	I32HEX FileType = 32
)
