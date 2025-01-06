package binary

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// FileFormatMagic is the magic string that needs to appear in the header.
const FileFormatMagic string = "ContainerSSH-Auditlog"

// FileFormatLength describes the length of the magic string. The non-used bytes need to be filled up with \000.
const FileFormatLength = 32

// CurrentVersion describes the current binary log version number
const CurrentVersion = uint64(2)

var fileFormatBytes []byte

// Header is the structure of the audit log header.
type Header struct {
	Magic   []byte
	Version uint64
}

func (h Header) getBytes() []byte {
	versionBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(versionBytes, h.Version)
	result := make([]byte, FileFormatLength+8)
	for i := 0; i < FileFormatLength; i++ {
		result[i] = fileFormatBytes[i]
	}
	for i := 0; i < 8; i++ {
		result[i+FileFormatLength] = versionBytes[i]
	}
	return result
}

func newHeader(version uint64) Header {
	return Header{
		Magic:   fileFormatBytes,
		Version: version,
	}
}

func readHeader(reader io.Reader, maxVersion uint64) (uint64, error) {
	headerBytes := make([]byte, FileFormatLength+8)

	_, err := io.ReadAtLeast(reader, headerBytes, FileFormatLength+8)
	if err != nil {
		return 0, err
	}
	if !bytes.Equal(headerBytes[:FileFormatLength], fileFormatBytes) {
		return 0, fmt.Errorf("invalid file format header: %v", headerBytes[:FileFormatLength])
	}
	version := binary.LittleEndian.Uint64(headerBytes[FileFormatLength:])
	if version > maxVersion {
		return 0, fmt.Errorf("file format version is higher than supported: %d", version)
	}
	return version, nil
}

func init() {
	fileFormatBytes = make([]byte, FileFormatLength)
	for i := 0; i < len(FileFormatMagic); i++ {
		fileFormatBytes[i] = FileFormatMagic[i]
	}
	for i := len(FileFormatMagic); i < FileFormatLength; i++ {
		fileFormatBytes[i] = "\000"[0]
	}
}
