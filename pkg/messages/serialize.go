package messages

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
)

func (m *Message) Serialize() ([]byte, error) {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %v", err)
	}

	compressed := bytes.NewBuffer(nil)
	gzipWriter := zlib.NewWriter(compressed)
	if _, err := gzipWriter.Write(jsonData); err != nil {
		return nil, fmt.Errorf("failed to compress message: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %v", err)
	}

	return compressed.Bytes(), nil
}

func DeserializeMessage(data []byte) (*Message, error) {
	gzipReader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decompress message: %v", err)
	}
	defer gzipReader.Close()

	jsonData, err := io.ReadAll(gzipReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read decompressed message: %v", err)
	}

	m := &Message{}
	if err := json.Unmarshal(jsonData, m); err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %v", err)
	}

	return m, nil
}
