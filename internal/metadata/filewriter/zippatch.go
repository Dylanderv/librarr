package filewriter

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"os"
	"time"
)

// Patching a zip in place, rather than rebuilding it, is what makes metadata
// edits on large archives usable. A 780 MB CBZ costs ~1.6 GB of I/O and a full
// decompress/recompress pass to rebuild; patching writes ~20 KB once.
//
// The trick is that a zip reader locates entries through the central directory
// at the end of the file, not by walking it from the start. Appending a new
// entry plus a fresh copy of the central directory therefore supersedes the old
// one without touching a single existing byte:
//
//	[ existing entry data ][ old central dir ][ new entry ][ new central dir ][ new EOCD ]
//	                        ^ dead weight, a few KB
//
// Because nothing is overwritten or truncated, a crash mid-write leaves the
// original archive readable: a reader scanning backwards finds either the new
// EOCD or the old one. That is at least as safe as the temp-file-and-rename
// path this replaces, and it is a single write syscall instead of two full
// passes over the file.
//
// The cost is a few KB of dead bytes per edit, which is not worth reclaiming.

const (
	sigLocalFile     = 0x04034b50
	sigCentralDir    = 0x02014b50
	sigEOCD          = 0x06054b50
	sigZip64EOCDLoc  = 0x07064b50
	eocdLen          = 22
	centralDirRecLen = 46
	localHeaderLen   = 30
	maxCommentLen    = 65535
	// uint32Max is the largest offset a non-ZIP64 archive can express.
	uint32Max = 0xffffffff
)

// errNotPatchable means the archive's layout is outside what patchZipEntry
// handles (ZIP64, a trailing EOCD comment, appended junk). Callers fall back to
// a full rewrite, which has no such constraints.
var errNotPatchable = errors.New("archive cannot be patched in place")

// patchZipEntry appends entry `name` with the given contents to the zip at
// path, superseding any existing entry of that name. The data is stored
// uncompressed: these are small XML documents, and Deflate would only add CPU.
func patchZipEntry(path, name string, data []byte) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	size := info.Size()

	oldCD, cdOffset, err := readCentralDir(f, size)
	if err != nil {
		return err
	}

	keptCD, keptCount, err := dropCentralDirRecord(oldCD, name)
	if err != nil {
		return err
	}

	// The new entry goes after the old central directory, so every existing
	// offset stays valid and its records can be reused verbatim.
	localOffset := size
	newCDOffset := localOffset + int64(localHeaderLen+len(name)+len(data))
	if newCDOffset > uint32Max || newCDOffset+int64(len(keptCD)) > uint32Max {
		// Past 4 GB the offsets no longer fit in the classic headers.
		return errNotPatchable
	}
	if cdOffset > uint32Max {
		return errNotPatchable
	}

	sum := crc32.ChecksumIEEE(data)
	dosTime, dosDate := dosTimestamp(info.ModTime())

	buf := make([]byte, 0, localHeaderLen+len(name)+len(data)+len(keptCD)+centralDirRecLen+len(name)+eocdLen)
	buf = appendLocalHeader(buf, name, data, sum, dosTime, dosDate)
	buf = append(buf, data...)
	buf = append(buf, keptCD...)
	buf = appendCentralDirRecord(buf, name, data, sum, dosTime, dosDate, uint32(localOffset))
	buf = appendEOCD(buf, keptCount+1, uint32(newCDOffset), uint32(len(keptCD)+centralDirRecLen+len(name)))

	// One write, at the end of the file. Nothing before `size` is touched.
	if _, err := f.WriteAt(buf, size); err != nil {
		return err
	}
	return f.Sync()
}

// readCentralDir returns the raw central directory bytes and its offset.
func readCentralDir(f *os.File, size int64) ([]byte, int64, error) {
	if size < eocdLen {
		return nil, 0, errNotPatchable
	}

	tailLen := int64(eocdLen + maxCommentLen)
	if tailLen > size {
		tailLen = size
	}
	tail := make([]byte, tailLen)
	if _, err := f.ReadAt(tail, size-tailLen); err != nil {
		return nil, 0, err
	}

	// Scan backwards for the last EOCD signature.
	pos := -1
	for i := len(tail) - eocdLen; i >= 0; i-- {
		if binary.LittleEndian.Uint32(tail[i:]) == sigEOCD {
			pos = i
			break
		}
	}
	if pos < 0 {
		return nil, 0, errNotPatchable
	}
	eocd := tail[pos:]

	// A comment, or anything after the EOCD, would end up buried by our append.
	if binary.LittleEndian.Uint16(eocd[20:]) != 0 || len(eocd) != eocdLen {
		return nil, 0, errNotPatchable
	}

	entries := binary.LittleEndian.Uint16(eocd[10:])
	cdSize := binary.LittleEndian.Uint32(eocd[12:])
	cdOffset := binary.LittleEndian.Uint32(eocd[16:])
	if entries == 0xffff || cdSize == uint32Max || cdOffset == uint32Max {
		return nil, 0, errNotPatchable // ZIP64
	}
	// A ZIP64 locator sits immediately before the EOCD when present.
	if pos >= 20 && binary.LittleEndian.Uint32(tail[pos-20:]) == sigZip64EOCDLoc {
		return nil, 0, errNotPatchable
	}
	if int64(cdOffset)+int64(cdSize) > size-eocdLen {
		return nil, 0, errNotPatchable
	}

	cd := make([]byte, cdSize)
	if _, err := f.ReadAt(cd, int64(cdOffset)); err != nil {
		return nil, 0, err
	}
	return cd, int64(cdOffset), nil
}

// dropCentralDirRecord returns the central directory with any record named
// `drop` removed, plus the number of records kept. Records are copied verbatim:
// their local-header offsets are untouched by an append, so re-encoding them
// would only risk losing fields we do not model.
func dropCentralDirRecord(cd []byte, drop string) ([]byte, uint16, error) {
	kept := make([]byte, 0, len(cd))
	count := 0

	for pos := 0; pos < len(cd); {
		if pos+centralDirRecLen > len(cd) {
			return nil, 0, errNotPatchable
		}
		if binary.LittleEndian.Uint32(cd[pos:]) != sigCentralDir {
			return nil, 0, errNotPatchable
		}
		nameLen := int(binary.LittleEndian.Uint16(cd[pos+28:]))
		extraLen := int(binary.LittleEndian.Uint16(cd[pos+30:]))
		commentLen := int(binary.LittleEndian.Uint16(cd[pos+32:]))
		recLen := centralDirRecLen + nameLen + extraLen + commentLen
		if pos+recLen > len(cd) {
			return nil, 0, errNotPatchable
		}

		if string(cd[pos+centralDirRecLen:pos+centralDirRecLen+nameLen]) != drop {
			kept = append(kept, cd[pos:pos+recLen]...)
			count++
		}
		pos += recLen
	}

	if count+1 > 0xffff {
		return nil, 0, errNotPatchable
	}
	return kept, uint16(count), nil
}

func appendLocalHeader(buf []byte, name string, data []byte, sum, dosTime, dosDate uint32) []byte {
	buf = binary.LittleEndian.AppendUint32(buf, sigLocalFile)
	buf = binary.LittleEndian.AppendUint16(buf, 20) // version needed
	buf = binary.LittleEndian.AppendUint16(buf, 0)  // flags
	buf = binary.LittleEndian.AppendUint16(buf, 0)  // method: stored
	buf = binary.LittleEndian.AppendUint16(buf, uint16(dosTime))
	buf = binary.LittleEndian.AppendUint16(buf, uint16(dosDate))
	buf = binary.LittleEndian.AppendUint32(buf, sum)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(data)))
	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(data)))
	buf = binary.LittleEndian.AppendUint16(buf, uint16(len(name)))
	buf = binary.LittleEndian.AppendUint16(buf, 0) // extra
	return append(buf, name...)
}

func appendCentralDirRecord(buf []byte, name string, data []byte, sum, dosTime, dosDate, localOffset uint32) []byte {
	buf = binary.LittleEndian.AppendUint32(buf, sigCentralDir)
	buf = binary.LittleEndian.AppendUint16(buf, 20) // version made by
	buf = binary.LittleEndian.AppendUint16(buf, 20) // version needed
	buf = binary.LittleEndian.AppendUint16(buf, 0)  // flags
	buf = binary.LittleEndian.AppendUint16(buf, 0)  // method: stored
	buf = binary.LittleEndian.AppendUint16(buf, uint16(dosTime))
	buf = binary.LittleEndian.AppendUint16(buf, uint16(dosDate))
	buf = binary.LittleEndian.AppendUint32(buf, sum)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(data)))
	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(data)))
	buf = binary.LittleEndian.AppendUint16(buf, uint16(len(name)))
	buf = binary.LittleEndian.AppendUint16(buf, 0)         // extra
	buf = binary.LittleEndian.AppendUint16(buf, 0)         // comment
	buf = binary.LittleEndian.AppendUint16(buf, 0)         // disk number
	buf = binary.LittleEndian.AppendUint16(buf, 0)         // internal attrs
	buf = binary.LittleEndian.AppendUint32(buf, 0o644<<16) // external attrs
	buf = binary.LittleEndian.AppendUint32(buf, localOffset)
	return append(buf, name...)
}

func appendEOCD(buf []byte, entries uint16, cdOffset, cdSize uint32) []byte {
	buf = binary.LittleEndian.AppendUint32(buf, sigEOCD)
	buf = binary.LittleEndian.AppendUint16(buf, 0) // this disk
	buf = binary.LittleEndian.AppendUint16(buf, 0) // disk with central dir
	buf = binary.LittleEndian.AppendUint16(buf, entries)
	buf = binary.LittleEndian.AppendUint16(buf, entries)
	buf = binary.LittleEndian.AppendUint32(buf, cdSize)
	buf = binary.LittleEndian.AppendUint32(buf, cdOffset)
	return binary.LittleEndian.AppendUint16(buf, 0) // comment length
}

// dosTimestamp encodes a time as the MS-DOS date/time pair zip headers use.
// Anything before 1980 is unrepresentable and clamps to the epoch's start.
func dosTimestamp(t time.Time) (dosTime, dosDate uint32) {
	if t.Year() < 1980 {
		return 0, 1<<5 | 1 // 1980-01-01 00:00
	}
	dosTime = uint32(t.Hour()<<11 | t.Minute()<<5 | t.Second()/2)
	dosDate = uint32((t.Year()-1980)<<9 | int(t.Month())<<5 | t.Day())
	return dosTime, dosDate
}
