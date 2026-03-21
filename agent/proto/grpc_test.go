package proto

import (
	"testing"

	"github.com/gogo/protobuf/proto"
)

// TestPackagedData_MarshalUnmarshal 测试 PackagedData 的序列化和反序列化
func TestPackagedData_MarshalUnmarshal(t *testing.T) {
	original := &PackagedData{
		Records: []*EncodedRecord{
			{
				DataType:  1001,
				Timestamp: 1234567890,
				Data:      []byte("test data"),
			},
		},
		AgentId:  "test-agent-id",
		Ipv4:     []string{"192.168.1.1", "10.0.0.1"},
		Hostname: "test-host",
		Version:  "1.0.0",
		Product:  "cloudsec-agent",
	}

	// 序列化
	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// 反序列化
	unmarshaled := &PackagedData{}
	if err := proto.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// 验证字段
	if unmarshaled.AgentId != original.AgentId {
		t.Errorf("AgentId mismatch: got %s, want %s", unmarshaled.AgentId, original.AgentId)
	}
	if unmarshaled.Hostname != original.Hostname {
		t.Errorf("Hostname mismatch: got %s, want %s", unmarshaled.Hostname, original.Hostname)
	}
	if len(unmarshaled.Records) != len(original.Records) {
		t.Errorf("Records length mismatch: got %d, want %d", len(unmarshaled.Records), len(original.Records))
	}
	if len(unmarshaled.Ipv4) != len(original.Ipv4) {
		t.Errorf("IPv4 length mismatch: got %d, want %d", len(unmarshaled.Ipv4), len(original.Ipv4))
	}
}

// TestRecord_MarshalUnmarshal 测试 Record 的序列化和反序列化
func TestRecord_MarshalUnmarshal(t *testing.T) {
	original := &Record{
		DataType:  1001,
		Timestamp: 1234567890,
		Data: &Payload{
			Fields: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	// 序列化
	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// 反序列化
	unmarshaled := &Record{}
	if err := proto.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// 验证字段
	if unmarshaled.DataType != original.DataType {
		t.Errorf("DataType mismatch: got %d, want %d", unmarshaled.DataType, original.DataType)
	}
	if unmarshaled.Timestamp != original.Timestamp {
		t.Errorf("Timestamp mismatch: got %d, want %d", unmarshaled.Timestamp, original.Timestamp)
	}
	if unmarshaled.Data == nil {
		t.Fatal("Data is nil")
	}
	if len(unmarshaled.Data.Fields) != len(original.Data.Fields) {
		t.Errorf("Fields length mismatch: got %d, want %d", len(unmarshaled.Data.Fields), len(original.Data.Fields))
	}
}

// TestCommand_MarshalUnmarshal 测试 Command 的序列化和反序列化
func TestCommand_MarshalUnmarshal(t *testing.T) {
	original := &Command{
		Ctrl: 1,
		Task: &Task{
			DataType:   1050,
			ObjectName: "agent",
			Data:       "test task data",
			Token:      "test-token",
		},
		Configs: []*Config{
			{
				Name:    "test-plugin",
				Type:    "collector",
				Version: "1.0.0",
				Sha256:  "abc123",
			},
		},
	}

	// 序列化
	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// 反序列化
	unmarshaled := &Command{}
	if err := proto.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// 验证字段
	if unmarshaled.Ctrl != original.Ctrl {
		t.Errorf("Ctrl mismatch: got %d, want %d", unmarshaled.Ctrl, original.Ctrl)
	}
	if unmarshaled.Task == nil {
		t.Fatal("Task is nil")
	}
	if unmarshaled.Task.ObjectName != original.Task.ObjectName {
		t.Errorf("Task.ObjectName mismatch: got %s, want %s", unmarshaled.Task.ObjectName, original.Task.ObjectName)
	}
	if len(unmarshaled.Configs) != len(original.Configs) {
		t.Errorf("Configs length mismatch: got %d, want %d", len(unmarshaled.Configs), len(original.Configs))
	}
}

// TestPayload_MapFields 测试 Payload 的 map 字段
func TestPayload_MapFields(t *testing.T) {
	original := &Payload{
		Fields: map[string]string{
			"field1": "value1",
			"field2": "value2",
			"field3": "value3",
		},
	}

	// 序列化
	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// 反序列化
	unmarshaled := &Payload{}
	if err := proto.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// 验证 map 字段
	if len(unmarshaled.Fields) != len(original.Fields) {
		t.Errorf("Fields length mismatch: got %d, want %d", len(unmarshaled.Fields), len(original.Fields))
	}
	for k, v := range original.Fields {
		if unmarshaled.Fields[k] != v {
			t.Errorf("Fields[%s] mismatch: got %s, want %s", k, unmarshaled.Fields[k], v)
		}
	}
}
