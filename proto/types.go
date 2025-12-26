package proto

import "math/bits"

// Config 插件配置
type Config struct {
	Name         string
	Type         string
	Version      string
	Sha256       string
	Signature    string
	DownloadUrls []string
	Detail       string
}

// Task 任务结构
type Task struct {
	DataType   int32
	ObjectName string
	Data       string
	Token      string
}

// EncodedRecord 编码后的记录
type EncodedRecord struct {
	DataType  int32
	Timestamp int64
	Data      []byte
}

// Size 计算 Task 序列化后的大小（protobuf wire format）
func (t *Task) Size() int {
	if t == nil {
		return 0
	}
	n := 0
	if t.DataType != 0 {
		n += 1 + sizeVarint(uint64(t.DataType))
	}
	if len(t.ObjectName) > 0 {
		n += 1 + len(t.ObjectName) + sizeVarint(uint64(len(t.ObjectName)))
	}
	if len(t.Data) > 0 {
		n += 1 + len(t.Data) + sizeVarint(uint64(len(t.Data)))
	}
	if len(t.Token) > 0 {
		n += 1 + len(t.Token) + sizeVarint(uint64(len(t.Token)))
	}
	return n
}

// MarshalToSizedBuffer 将 Task 序列化到缓冲区（protobuf wire format）
func (t *Task) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	if t == nil {
		return 0, nil
	}
	i := len(dAtA)

	// Token (field 4)
	if len(t.Token) > 0 {
		i -= len(t.Token)
		copy(dAtA[i:], t.Token)
		i = encodeVarint(dAtA, i, uint64(len(t.Token)))
		i--
		dAtA[i] = 0x22 // field number 4, wire type 2 (length-delimited)
	}

	// Data (field 3)
	if len(t.Data) > 0 {
		i -= len(t.Data)
		copy(dAtA[i:], t.Data)
		i = encodeVarint(dAtA, i, uint64(len(t.Data)))
		i--
		dAtA[i] = 0x1a // field number 3, wire type 2
	}

	// ObjectName (field 2)
	if len(t.ObjectName) > 0 {
		i -= len(t.ObjectName)
		copy(dAtA[i:], t.ObjectName)
		i = encodeVarint(dAtA, i, uint64(len(t.ObjectName)))
		i--
		dAtA[i] = 0x12 // field number 2, wire type 2
	}

	// DataType (field 1)
	if t.DataType != 0 {
		i = encodeVarint(dAtA, i, uint64(t.DataType))
		i--
		dAtA[i] = 0x8 // field number 1, wire type 0 (varint)
	}

	return len(dAtA) - i, nil
}

// sizeVarint 计算 varint 编码后的大小
func sizeVarint(x uint64) int {
	return (bits.Len64(x|1) + 6) / 7
}

// encodeVarint 将 uint64 编码为 varint 格式
func encodeVarint(dAtA []byte, offset int, v uint64) int {
	offset -= sizeVarint(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
