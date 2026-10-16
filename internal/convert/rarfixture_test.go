package convert

// Hand-rolled RAR 4.x fixture generator: librarr runs anywhere (CI included)
// with no rar/unrar binary, and no third-party RAR writer exists in pure Go,
// so the tests build their own minimal stored-method archives instead of
// vendoring binaries.
//
// The layout was validated against the reader (nwaples/rardecode) and
// cross-checked with 7-Zip: a RAR 4 marker, a minimal main block, one file
// block per entry with stored data, and an end block. Only stored entries are
// generated — compression is the reader's job to decode, not the fixture's to
// encode.

import (
	"encoding/binary"
	"hash/crc32"
	"os"
	"testing"
	"time"
)

type rarEntry struct {
	name      string
	data      []byte
	dir       bool
	encrypted bool
}

func rarBlock(typ byte, flags uint16, payload []byte) []byte {
	size := 7 + len(payload)
	h := crc32.NewIEEE()
	var hdr [5]byte
	hdr[0] = typ
	binary.LittleEndian.PutUint16(hdr[1:3], flags)
	binary.LittleEndian.PutUint16(hdr[3:5], uint16(size))
	h.Write(hdr[:])
	h.Write(payload)
	var out []byte
	var tmp [2]byte
	binary.LittleEndian.PutUint16(tmp[:], uint16(h.Sum32()))
	out = append(out, tmp[:]...)
	out = append(out, typ)
	binary.LittleEndian.PutUint16(tmp[:], flags)
	out = append(out, tmp[:]...)
	binary.LittleEndian.PutUint16(tmp[:], uint16(size))
	out = append(out, tmp[:]...)
	out = append(out, payload...)
	return out
}

// attr is the DOS attribute byte: 0x20 for files, 0x10 for directories —
// readers like 7-Zip tell the two apart by it.
func rarFileHead(pack, unp uint32, name string, mod uint32, withPack bool, attr uint32) []byte {
	var p []byte
	var tmp [4]byte
	put := func(v uint32) { binary.LittleEndian.PutUint32(tmp[:], v); p = append(p, tmp[:]...) }
	if withPack {
		put(pack)
	}
	put(unp)
	p = append(p, 0x00) // HOST_OS
	put(0)              // FILE_CRC placeholder, fixed by caller for files
	p = append(p, 0x00, 0x00, 0x00, 0x00)
	copy(p[len(p)-4:], binary.LittleEndian.AppendUint32(nil, mod))
	p = append(p, 0x1D) // UNP_VER
	p = append(p, 0x30) // METHOD stored
	var ns [2]byte
	binary.LittleEndian.PutUint16(ns[:], uint16(len(name)))
	p = append(p, ns[:]...)
	put(attr) // ATTR
	p = append(p, name...)
	return p
}

func dosTimestamp() uint32 {
	t := time.Date(2024, 5, 6, 7, 8, 10, 0, time.UTC)
	return uint32(t.Second()/2) | uint32(t.Minute())<<5 | uint32(t.Hour())<<11 |
		uint32(t.Day())<<16 | uint32(t.Month())<<21 | uint32(t.Year()-1980)<<25
}

func writeStoredRAR(t *testing.T, path string, entries []rarEntry) {
	t.Helper()
	writeStoredRARVariant(t, path, entries, []byte{0x00, 0x00}, true)
}

func writeStoredRARVariant(t *testing.T, path string, entries []rarEntry, mainPayload []byte, withEnd bool) {
	t.Helper()
	var out []byte
	out = append(out, 'R', 'a', 'r', '!', 0x1A, 0x07, 0x00)
	// Real RAR 4 archives carry a 13-byte main block (HIGH_POSAV and friends);
	// readers like 7-Zip refuse anything else, so the fixture matches that.
	out = append(out, rarBlock(0x73, 0x0000, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00})...)
	mod := dosTimestamp()
	for _, e := range entries {
		if e.dir {
			// WinRAR stores folders as file blocks with zero sizes and the
			// data flag set (PACK_SIZE present but empty), not as dataless
			// blocks — readers like 7-Zip reject the dataless form.
			head := rarFileHead(0, 0, e.name, mod, true, 0x10)
			out = append(out, rarBlock(0x74, 0x80E0, head)...)
			continue
		}
		head := rarFileHead(uint32(len(e.data)), uint32(len(e.data)), e.name, mod, true, 0x20)
		copy(head[9:13], binary.LittleEndian.AppendUint32(nil, crc32.ChecksumIEEE(e.data)))
		flags := uint16(0x8000)
		if e.encrypted {
			flags |= 0x0004
		}
		out = append(out, rarBlock(0x74, flags, head)...)
		out = append(out, e.data...)
	}
	if withEnd {
		out = append(out, rarBlock(0x7B, 0x0000, nil)...)
	}
	if err := os.WriteFile(path, out, 0o666); err != nil {
		t.Fatal(err)
	}
}
