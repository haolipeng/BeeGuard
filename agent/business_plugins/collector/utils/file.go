package utils

import (
	"bufio"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"sync"
)

var (
	mp = sync.Pool{
		New: func() interface{} { return md5.New() },
	}
	hp = sync.Pool{
		New: func() interface{} { return &xxhash{} },
	}
	rp = sync.Pool{
		New: func() interface{} {
			return bufio.NewReaderSize(nil, 1024*1024)
		},
	}
)

// xxhash 简化版本，用于文件哈希计算
type xxhash struct {
	h uint64
}

func (x *xxhash) Write(p []byte) (n int, err error) {
	for _, b := range p {
		x.h = x.h*0x9e3779b185ebca87 + uint64(b)
	}
	return len(p), nil
}

func (x *xxhash) Sum(b []byte) []byte {
	s := make([]byte, 8)
	binary.LittleEndian.PutUint64(s, x.h)
	if b == nil {
		return s
	}
	return append(b, s...)
}

func (x *xxhash) Reset() {
	x.h = 0
}

func (x *xxhash) Size() int {
	return 8
}

func (x *xxhash) BlockSize() int {
	return 1
}

func caculateMd5(f *os.File) (ret string, err error) {
	r := rp.Get().(*bufio.Reader)
	defer r.Reset(nil)
	defer rp.Put(r)
	h := mp.Get().(hash.Hash)
	defer h.Reset()
	defer mp.Put(h)
	var s fs.FileInfo
	s, err = f.Stat()
	if err != nil {
		return
	}
	if s.Size() > 100*1024*1024 {
		err = fmt.Errorf("file size is larger than limitation: %v", s.Size())
		return
	}
	r.Reset(f)
	_, err = io.Copy(h, r)
	if err != nil {
		return
	}
	ret = hex.EncodeToString(h.Sum(nil))
	return
}

func GetMd5(path string, procPath string) (ret string, err error) {
	var f *os.File
	f, err = os.Open(path)
	if err == nil {
		ret, err = caculateMd5(f)
		f.Close()
		return
	}
	if procPath == "" {
		return
	}
	f, err = os.Open(procPath)
	if err != nil {
		return
	}
	ret, err = caculateMd5(f)
	f.Close()
	return
}

func caculateHash(f *os.File) (ret string, err error) {
	r := rp.Get().(*bufio.Reader)
	defer r.Reset(nil)
	defer rp.Put(r)
	h := hp.Get().(hash.Hash)
	defer h.Reset()
	defer hp.Put(h)
	var s fs.FileInfo
	s, err = f.Stat()
	if err != nil {
		return
	}
	err = binary.Write(h, binary.LittleEndian, uint64(s.Size()))
	if err != nil {
		return
	}
	r.Reset(f)
	lr := io.LimitReader(r, 32*1024)
	_, err = io.Copy(h, lr)
	if err != nil {
		return
	}
	ret = hex.EncodeToString(h.Sum(nil))
	return
}

func GetHash(path string, procPath string) (ret string, err error) {
	var f *os.File
	f, err = os.Open(path)
	if err == nil {
		ret, err = caculateHash(f)
		f.Close()
		return
	}
	if procPath == "" {
		return
	}
	f, err = os.Open(procPath)
	if err != nil {
		return
	}
	defer f.Close()
	ret, err = caculateHash(f)
	return
}
