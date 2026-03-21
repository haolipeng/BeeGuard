package buffer

import (
	"sync"

	"gitlab.myinterest.top/security/agent/proto"
)

var (
	pool = sync.Pool{
		New: func() interface{} {
			return &proto.EncodedRecord{
				Data: make([]byte, 0, 1024*2),
			}
		},
	}
)

// 单次从池中获取
func GetEncodedRecord() *proto.EncodedRecord {
	return pool.Get().(*proto.EncodedRecord)
}

// 单次放回池中
func PutEncodedRecord(rec *proto.EncodedRecord) {
	pool.Put(rec)
}

// 批量放回池中
func PutEncodedRecords(recs []*proto.EncodedRecord) {
	for _, rec := range recs {
		rec.Data = rec.Data[:0]
		pool.Put(rec)
	}
}
