package sauce

import (
	"encoding/binary"
	"io"
	"os"
)

const (
	RecordSize = 128
	ID         = "SAUCE"
	CommentID  = "COMNT"
)

// Record holds the metadata from a SAUCE record.
type Record struct {
	Version      string // 2 bytes
	Title        string // 35 bytes
	Author       string // 20 bytes
	Group        string // 20 bytes
	Date         string // 8 bytes
	FileSize     uint32 // 4 bytes
	DataType     uint8  // 1 byte
	FileType     uint8  // 1 byte
	TInfo1       uint16 // 2 bytes
	TInfo2       uint16 // 2 bytes
	TInfo3       uint16 // 2 bytes
	TInfo4       uint16 // 2 bytes
	Comments     uint8  // 1 byte
	Flags        uint8  // 1 byte
	TInfoS       string // 22 bytes
	CommentLines []string
}

// Get finds and parses a SAUCE record from the end of a file.
func Get(f *os.File) (*Record, error) {
	// A SAUCE record is at the end of the file, so we seek backwards.
	// We first look for the record itself, then for an optional comment block.
	offset, err := f.Seek(-int64(RecordSize), io.SeekEnd)
	if err != nil {
		// File is likely too small to contain a SAUCE record.
		return nil, nil
	}

	buf := make([]byte, RecordSize)
	if _, err := io.ReadFull(f, buf); err != nil {
		return nil, err
	}

	if string(buf[0:5]) != ID {
		return nil, nil // No SAUCE record found.
	}

	rec := &Record{
		Version:  string(buf[5:7]),
		Title:    string(buf[7:42]),
		Author:   string(buf[42:62]),
		Group:    string(buf[62:82]),
		Date:     string(buf[82:90]),
		FileSize: binary.LittleEndian.Uint32(buf[90:94]),
		DataType: buf[94],
		FileType: buf[95],
		TInfo1:   binary.LittleEndian.Uint16(buf[96:98]),
		TInfo2:   binary.LittleEndian.Uint16(buf[98:100]),
		TInfo3:   binary.LittleEndian.Uint16(buf[100:102]),
		TInfo4:   binary.LittleEndian.Uint16(buf[102:104]),
		Comments: buf[104],
		Flags:    buf[105],
		TInfoS:   string(buf[106:128]),
	}

	// If there are comments, read them.
	if rec.Comments > 0 {
		commentBlockSize := int64(rec.Comments) * 64
		// Seek to the beginning of the comment block
		commentOffset := offset - 5 - commentBlockSize
		if _, err := f.Seek(commentOffset, io.SeekStart); err != nil {
			return rec, nil // Return record without comments if seek fails
		}

		commentBuf := make([]byte, 5)
		if _, err := io.ReadFull(f, commentBuf); err != nil || string(commentBuf) != CommentID {
			return rec, nil // Comment block not found where expected.
		}

		// Read the actual comments
		rec.CommentLines = make([]string, rec.Comments)
		for i := 0; i < int(rec.Comments); i++ {
			lineBuf := make([]byte, 64)
			if _, err := io.ReadFull(f, lineBuf); err != nil {
				break
			}
			rec.CommentLines[i] = string(lineBuf)
		}
	}
	return rec, nil
}
