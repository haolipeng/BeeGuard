package buffer

import (
	"errors"
	"sync"

	"github.com/haolipeng/BeeGuard/agent/proto"
)

var (
	mu                = &sync.Mutex{}
	buf               = [8192]*proto.EncodedRecord{}
	offset            = 0
	ErrbufferOverflow = errors.New("buffer overflow")
)

func WriteEncodedRecord(rec *proto.EncodedRecord) (err error) {
	mu.Lock()
	if offset < len(buf) {
		buf[offset] = rec
		offset++
	}
	mu.Unlock()
	return
}

func ReadEncodedRecords() (ret []*proto.EncodedRecord) {
	mu.Lock()
	ret = make([]*proto.EncodedRecord, offset)
	copy(ret, buf[:offset])
	offset = 0
	mu.Unlock()
	return
}
