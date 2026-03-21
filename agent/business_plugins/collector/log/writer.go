package log

import (
	"encoding/json"
	"strconv"
	"time"

	businessplugins "business_plugins/lib"
)

type remoteWriter struct {
	client *businessplugins.Client
}

func (w *remoteWriter) Write(p []byte) (n int, err error) {
	if w.client != nil {
		rec := &businessplugins.Record{
			DataType: 1011,
			Data: &businessplugins.Payload{
				Fields: map[string]string{},
			},
		}
		fields := map[string]interface{}{}
		err = json.Unmarshal(p, &fields)
		if err != nil {
			return
		}
		timestamp, ok := fields["timestamp"]
		if ok {
			timestamp, err := strconv.ParseInt(timestamp.(string), 10, 64)
			if err == nil {
				rec.Timestamp = timestamp
				delete(fields, "timestamp")
			}
		}
		if rec.Timestamp == 0 {
			rec.Timestamp = time.Now().Unix()
		}
		for k, v := range fields {
			switch v := v.(type) {
			case string:
				rec.Data.Fields[k] = v
			case int:
				rec.Data.Fields[k] = strconv.Itoa(v)
			}
		}
		err = w.client.SendRecord(rec)
		if err != nil {
			return
		}
	}
	n = len(p)
	return
}

func (w *remoteWriter) Sync() error {
	if w.client != nil {
		return w.client.Flush()
	} else {
		return nil
	}
}
