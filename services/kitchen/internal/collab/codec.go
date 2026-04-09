package collab

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCodec(jsonCodec{})
}

// jsonCodec replaces the default protobuf codec with JSON so that EditRequest
// and EditResponse can be plain Go structs without a protoc dependency.
type jsonCodec struct{}

func (jsonCodec) Name() string { return "proto" }

func (jsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
