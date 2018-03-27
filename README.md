# IHEX - Go reader / writer for Intel HEX files

Library for reading and writing Intel HEX files according to the specification outlined in [Wikipedia's Article](https://en.wikipedia.org/wiki/Intel_HEX).

## Features

* Reading and Writing HEX files
* HEX Record Types
* Automatic HEX record checksum creation / validation
* Automatic HEX record type validation against HEX file type specifications
* Supports I8HEX, I16HEX, and I32HEX file specifications

## Installation

Add the littlehawk93/ihex package to your project using the standard go import method:

    go get github.com/littlehawk93/ihex

Once installed, include the package in your project with the following import command:

    import "github.com/littlehawk93/ihex"

## Examples

To read an Intel HEX file in the I8HEX specification:

    file, err := os.Open("/path/to/hex/file.hex")

    if err == nil {
        
        hexFile := ihex.NewHexFile()

        hexFile.Type = ihex.HexFileTypeI8HEX

        _, err := hexFile.ReadFrom(file)

        if err == nil {

            // Use hexFile here
        }
    }

Once parsed, a HEX file struct can be read programmatically to get data from it:

    // Iterate over all records in HEX file
    for hexFile.Next() {

        record := hexFile.Record()

        type := record.RecordType
        address := record.AddressOffset

        data := record.Data

        // Do stuff
    }

You can begin reading a HEX file from the beginning again at any time by calling Reset():

    hexFile.Reset()

You can also programmatically construct your own HEX file by adding records:

    hexFile := ihex.NewHexFile()
    hexFile.Type = ihex.HexFileTypeI16HEX

    var record ihex.Record

    record.RecordType = ihex.RecordTypeEndOfFile
    record.AddressOffset = 0
    record.Data = make([]byte, 0)

    hexFile.AddRecord(&record)

HEX file structs can be written back to Intel HEX file specification to any writer interface like so:

    file, err := os.Create("/path/to/hex/file.hex")

    if err == nil {
        
        _, err = hexFile.WriteTo(file)

        // Error handling
    }
