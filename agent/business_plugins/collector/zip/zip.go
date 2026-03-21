package zip

import (
	"archive/zip"
	"bufio"
	"compress/flate"
	"encoding/binary"
	"errors"
	"hash"
	"hash/crc32"
	"io"
	"io/fs"
	"os"
	"unicode/utf8"
)

var (
	ErrFormat    = errors.New("zip: not a valid zip file")
	ErrAlgorithm = errors.New("zip: unsupported compression algorithm")
	ErrChecksum  = errors.New("zip: checksum error")
)

const (
	fileHeaderSignature      = 0x04034b50
	directoryHeaderSignature = 0x02014b50
	directoryEndSignature    = 0x06054b50
	directory64LocSignature  = 0x07064b50
	directory64EndSignature  = 0x06064b50
	dataDescriptorSignature  = 0x08074b50
	fileHeaderLen            = 30
	directoryHeaderLen       = 46
	directoryEndLen          = 22
	dataDescriptorLen        = 16
	dataDescriptor64Len      = 24
	directory64LocLen        = 20
	directory64EndLen        = 56

	uint16max = (1 << 16) - 1
	uint32max = (1 << 32) - 1

	zip64ExtraID       = 0x0001
	ntfsExtraID        = 0x000a
	unixExtraID        = 0x000d
	extTimeExtraID     = 0x5455
	infoZipUnixExtraID = 0x5855
)

type readBuf []byte

func (b *readBuf) uint16() uint16 {
	v := binary.LittleEndian.Uint16(*b)
	*b = (*b)[2:]
	return v
}

func (b *readBuf) uint32() uint32 {
	v := binary.LittleEndian.Uint32(*b)
	*b = (*b)[4:]
	return v
}

func (b *readBuf) uint64() uint64 {
	v := binary.LittleEndian.Uint64(*b)
	*b = (*b)[8:]
	return v
}

func (b *readBuf) sub(n int) readBuf {
	b2 := (*b)[:n]
	*b = (*b)[n:]
	return b2
}

type checksumReader struct {
	rc    io.ReadCloser
	hash  hash.Hash32
	nread uint64
	f     *File
	err   error
}

func (r *checksumReader) Read(b []byte) (n int, err error) {
	if r.err != nil {
		return 0, r.err
	}
	n, err = r.rc.Read(b)
	r.hash.Write(b[:n])
	r.nread += uint64(n)
	if err == nil {
		return
	}
	if err == io.EOF {
		if r.nread != r.f.UncompressedSize64 {
			return 0, io.ErrUnexpectedEOF
		}
		if r.f.hasDataDescriptor() {
			if r.f.descErr != nil {
				if r.f.descErr == io.EOF {
					err = io.ErrUnexpectedEOF
				} else {
					err = r.f.descErr
				}
			} else if r.hash.Sum32() != r.f.CRC32 {
				err = ErrChecksum
			}
		} else {
			if r.f.CRC32 != 0 && r.hash.Sum32() != r.f.CRC32 {
				err = ErrChecksum
			}
		}
	}
	r.err = err
	return
}

func (r *checksumReader) Close() error { return r.rc.Close() }

type directoryEnd struct {
	diskNbr            uint32
	dirDiskNbr         uint32
	dirRecordsThisDisk uint64
	directoryRecords   uint64
	directorySize      uint64
	directoryOffset    uint64
	commentLen         uint16
	comment            string
}

type Reader struct {
	f             *os.File
	br            *bufio.Reader
	decompressors map[uint16]zip.Decompressor
}

type File struct {
	FileHeader
	zip          *Reader
	zipr         io.ReaderAt
	headerOffset int64
	zip64        bool
	descErr      error
}

type FileHeader struct {
	Name               string
	NonUTF8            bool
	CreatorVersion     uint16
	ReaderVersion      uint16
	Flags              uint16
	Method             uint16
	CRC32              uint32
	CompressedSize     uint32
	UncompressedSize   uint32
	CompressedSize64   uint64
	UncompressedSize64 uint64
	Extra              []byte
}

func FileInfoHeader(fi fs.FileInfo) (*FileHeader, error) {
	size := fi.Size()
	fh := &FileHeader{
		Name:               fi.Name(),
		UncompressedSize64: uint64(size),
	}
	if fh.UncompressedSize64 > uint32max {
		fh.UncompressedSize = uint32max
	} else {
		fh.UncompressedSize = uint32(fh.UncompressedSize64)
	}
	return fh, nil
}

type WalkFunc func(*File)

func OpenReader(name string) (r *Reader, err error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	size := fi.Size()
	var end *directoryEnd
	end, err = readDirectoryEnd(f, size)
	if err != nil {
		return
	}
	rs := io.NewSectionReader(f, 0, size)
	_, err = rs.Seek(int64(end.directoryOffset), io.SeekStart)
	if err != nil {
		return
	}
	return &Reader{
		f:  f,
		br: bufio.NewReader(rs),
		decompressors: map[uint16]zip.Decompressor{
			zip.Store:   io.NopCloser,
			zip.Deflate: flate.NewReader,
		},
	}, nil
}

func (r *Reader) WalkFiles(fn WalkFunc) (err error) {
	for {
		f := &File{zip: r, zipr: r.f}
		err = readDirectoryHeader(f, r.br)
		if err == ErrFormat || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return err
		}
		f.readDataDescriptor()
		fn(f)
	}
	return
}

func (r *Reader) Close() {
	r.f.Close()
}

func (r *Reader) Name() string {
	return r.f.Name()
}

func (h *FileHeader) isZip64() bool {
	return h.CompressedSize64 >= uint32max || h.UncompressedSize64 >= uint32max
}

func (f *FileHeader) hasDataDescriptor() bool {
	return f.Flags&0x8 != 0
}

func (f *File) findBodyOffset() (int64, error) {
	var buf [fileHeaderLen]byte
	if _, err := f.zipr.ReadAt(buf[:], f.headerOffset); err != nil {
		return 0, err
	}
	b := readBuf(buf[:])
	if sig := b.uint32(); sig != fileHeaderSignature {
		return 0, ErrFormat
	}
	b = b[22:]
	filenameLen := int(b.uint16())
	extraLen := int(b.uint16())
	return int64(fileHeaderLen + filenameLen + extraLen), nil
}

func (f *File) DataOffset() (offset int64, err error) {
	bodyOffset, err := f.findBodyOffset()
	if err != nil {
		return
	}
	return f.headerOffset + bodyOffset, nil
}

func (f *File) Open() (io.ReadCloser, error) {
	bodyOffset, err := f.findBodyOffset()
	if err != nil {
		return nil, err
	}
	size := int64(f.CompressedSize64)
	r := io.NewSectionReader(f.zipr, f.headerOffset+bodyOffset, size)
	dcomp := f.zip.decompressors[f.Method]
	if dcomp == nil {
		return nil, ErrAlgorithm
	}
	var rc io.ReadCloser = dcomp(r)
	rc = &checksumReader{
		rc:   rc,
		hash: crc32.NewIEEE(),
		f:    f,
	}
	return rc, nil
}

func (f *File) OpenRaw() (io.Reader, error) {
	bodyOffset, err := f.findBodyOffset()
	if err != nil {
		return nil, err
	}
	r := io.NewSectionReader(f.zipr, f.headerOffset+bodyOffset, int64(f.CompressedSize64))
	return r, nil
}

func (f *File) readDataDescriptor() {
	if !f.hasDataDescriptor() {
		return
	}

	bodyOffset, err := f.findBodyOffset()
	if err != nil {
		f.descErr = err
		return
	}

	zip64 := f.zip64 || f.isZip64()
	n := int64(dataDescriptorLen)
	if zip64 {
		n = dataDescriptor64Len
	}
	size := int64(f.CompressedSize64)
	r := io.NewSectionReader(f.zipr, f.headerOffset+bodyOffset+size, n)
	dd, err := readDataDescriptor(r, zip64)
	if err != nil {
		f.descErr = err
		return
	}
	f.CRC32 = dd.crc32
}

func findSignatureInBlock(b []byte) int {
	for i := len(b) - directoryEndLen; i >= 0; i-- {
		if b[i] == 'P' && b[i+1] == 'K' && b[i+2] == 0x05 && b[i+3] == 0x06 {
			n := int(b[i+directoryEndLen-2]) | int(b[i+directoryEndLen-1])<<8
			if n+directoryEndLen+i <= len(b) {
				return i
			}
		}
	}
	return -1
}

func readDirectoryEnd(r io.ReaderAt, size int64) (dir *directoryEnd, err error) {
	var buf []byte
	var directoryEndOffset int64
	for i, bLen := range []int64{1024, 65 * 1024} {
		if bLen > size {
			bLen = size
		}
		buf = make([]byte, int(bLen))
		if _, err := r.ReadAt(buf, size-bLen); err != nil && err != io.EOF {
			return nil, err
		}
		if p := findSignatureInBlock(buf); p >= 0 {
			buf = buf[p:]
			directoryEndOffset = size - bLen + int64(p)
			break
		}
		if i == 1 || bLen == size {
			return nil, ErrFormat
		}
	}

	b := readBuf(buf[4:])
	d := &directoryEnd{
		diskNbr:            uint32(b.uint16()),
		dirDiskNbr:         uint32(b.uint16()),
		dirRecordsThisDisk: uint64(b.uint16()),
		directoryRecords:   uint64(b.uint16()),
		directorySize:      uint64(b.uint32()),
		directoryOffset:    uint64(b.uint32()),
		commentLen:         b.uint16(),
	}
	l := int(d.commentLen)
	if l > len(b) {
		return nil, errors.New("zip: invalid comment length")
	}
	d.comment = string(b[:l])

	if d.directoryRecords == 0xffff || d.directorySize == 0xffff || d.directoryOffset == 0xffffffff {
		p, err := findDirectory64End(r, directoryEndOffset)
		if err == nil && p >= 0 {
			err = readDirectory64End(r, p, d)
		}
		if err != nil {
			return nil, err
		}
	}
	if o := int64(d.directoryOffset); o < 0 || o >= size {
		return nil, ErrFormat
	}
	return d, nil
}

func findDirectory64End(r io.ReaderAt, directoryEndOffset int64) (int64, error) {
	locOffset := directoryEndOffset - directory64LocLen
	if locOffset < 0 {
		return -1, nil
	}
	buf := make([]byte, directory64LocLen)
	if _, err := r.ReadAt(buf, locOffset); err != nil {
		return -1, err
	}
	b := readBuf(buf)
	if sig := b.uint32(); sig != directory64LocSignature {
		return -1, nil
	}
	if b.uint32() != 0 {
		return -1, nil
	}
	p := b.uint64()
	if b.uint32() != 1 {
		return -1, nil
	}
	return int64(p), nil
}

func readDirectory64End(r io.ReaderAt, offset int64, d *directoryEnd) (err error) {
	buf := make([]byte, directory64EndLen)
	if _, err := r.ReadAt(buf, offset); err != nil {
		return err
	}

	b := readBuf(buf)
	if sig := b.uint32(); sig != directory64EndSignature {
		return ErrFormat
	}

	b = b[12:]
	d.diskNbr = b.uint32()
	d.dirDiskNbr = b.uint32()
	d.dirRecordsThisDisk = b.uint64()
	d.directoryRecords = b.uint64()
	d.directorySize = b.uint64()
	d.directoryOffset = b.uint64()

	return nil
}

func readDirectoryHeader(f *File, r io.Reader) error {
	var buf [directoryHeaderLen]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return err
	}
	b := readBuf(buf[:])
	if sig := b.uint32(); sig != directoryHeaderSignature {
		return ErrFormat
	}
	f.CreatorVersion = b.uint16()
	f.ReaderVersion = b.uint16()
	f.Flags = b.uint16()
	f.Method = b.uint16()
	_ = b.uint16()
	_ = b.uint16()
	f.CRC32 = b.uint32()
	f.CompressedSize = b.uint32()
	f.UncompressedSize = b.uint32()
	f.CompressedSize64 = uint64(f.CompressedSize)
	f.UncompressedSize64 = uint64(f.UncompressedSize)
	filenameLen := int(b.uint16())
	extraLen := int(b.uint16())
	commentLen := int(b.uint16())
	b = b[4:]
	_ = b.uint32()
	f.headerOffset = int64(b.uint32())
	d := make([]byte, filenameLen+extraLen+commentLen)
	if _, err := io.ReadFull(r, d); err != nil {
		return err
	}
	f.Name = string(d[:filenameLen])
	f.Extra = d[filenameLen : filenameLen+extraLen]

	utf8Valid1, utf8Require1 := detectUTF8(f.Name)
	switch {
	case !utf8Valid1:
		f.NonUTF8 = true
	case !utf8Require1:
		f.NonUTF8 = false
	default:
		f.NonUTF8 = f.Flags&0x800 == 0
	}

	needUSize := f.UncompressedSize == ^uint32(0)
	needCSize := f.CompressedSize == ^uint32(0)
	needHeaderOffset := f.headerOffset == int64(^uint32(0))
	for extra := readBuf(f.Extra); len(extra) >= 4; {
		fieldTag := extra.uint16()
		fieldSize := int(extra.uint16())
		if len(extra) < fieldSize {
			break
		}
		fieldBuf := extra.sub(fieldSize)

		switch fieldTag {
		case zip64ExtraID:
			f.zip64 = true

			if needUSize {
				needUSize = false
				if len(fieldBuf) < 8 {
					return ErrFormat
				}
				f.UncompressedSize64 = fieldBuf.uint64()
			}
			if needCSize {
				needCSize = false
				if len(fieldBuf) < 8 {
					return ErrFormat
				}
				f.CompressedSize64 = fieldBuf.uint64()
			}
			if needHeaderOffset {
				needHeaderOffset = false
				if len(fieldBuf) < 8 {
					return ErrFormat
				}
				f.headerOffset = int64(fieldBuf.uint64())
			}
		}
	}
	_ = needUSize

	if needCSize || needHeaderOffset {
		return ErrFormat
	}

	return nil
}

func readDataDescriptor(r io.Reader, zip64 bool) (*dataDescriptor, error) {
	var buf [dataDescriptor64Len]byte

	if _, err := io.ReadFull(r, buf[:4]); err != nil {
		return nil, err
	}
	off := 0
	maybeSig := readBuf(buf[:4])
	if maybeSig.uint32() != dataDescriptorSignature {
		off += 4
	}

	end := dataDescriptorLen - 4
	if zip64 {
		end = dataDescriptor64Len - 4
	}
	if _, err := io.ReadFull(r, buf[off:end]); err != nil {
		return nil, err
	}
	b := readBuf(buf[:end])

	out := &dataDescriptor{
		crc32: b.uint32(),
	}

	if zip64 {
		out.compressedSize = b.uint64()
		out.uncompressedSize = b.uint64()
	} else {
		out.compressedSize = uint64(b.uint32())
		out.uncompressedSize = uint64(b.uint32())
	}
	return out, nil
}

func detectUTF8(s string) (valid, require bool) {
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		if r < 0x20 || r > 0x7d || r == 0x5c {
			if !utf8.ValidRune(r) || (r == utf8.RuneError && size == 1) {
				return false, false
			}
			require = true
		}
	}
	return true, require
}

type dataDescriptor struct {
	crc32            uint32
	compressedSize   uint64
	uncompressedSize uint64
}
