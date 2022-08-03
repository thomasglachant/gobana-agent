package core

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

const SyncLogin = "spooter"

type LogMetadata struct {
	AgentVersion string    `json:"agent_version" yaml:"agent_version"`
	Application  string    `json:"application" yaml:"application"`
	Server       string    `json:"server" yaml:"server"`
	Filename     string    `json:"filename" yaml:"filename"`
	Parser       string    `json:"parser" yaml:"parser"`
	CaptureDate  time.Time `json:"capture_date" yaml:"capture_date"`
}

type LogLine struct {
	Metadata LogMetadata       `yaml:"metadata" json:"metadata"`
	Date     time.Time         `yaml:"date" json:"date"`
	Raw      string            `yaml:"raw" json:"raw"`
	Fields   map[string]string `yaml:"fields" json:"fields"`
}

type SynchronizeLogsMessage struct {
	Logs []*LogLine `json:"logs"`
}

func EncryptMessage(data any, key string) ([]byte, error) {
	// marshal message
	message, _ := json.Marshal(data)

	// compress message
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err := gzipWriter.Write(message); err != nil {
		return nil, fmt.Errorf("error compressing message: %s", err)
	}
	err := gzipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing compress message: %s", err)
	}

	// encrypt message
	encryptedData, err := AESEncrypt(buf.Bytes(), key)
	if err != nil {
		return nil, fmt.Errorf("error encrypting data: %s", err)
	}

	return encryptedData, nil
}

func DecryptMessage(data []byte, key string, obj any) error {
	// decrypt content
	var decryptedBody []byte
	var err error
	decryptedBody, err = AESDecrypt(data, key)
	if err != nil {
		return fmt.Errorf("error decrypting data: %s", err)
	}

	// decompress message
	rdata := bytes.NewReader(decryptedBody)
	r, _ := gzip.NewReader(rdata)
	decryptedBody, _ = io.ReadAll(r)

	// unmarshal message
	if err := json.Unmarshal(decryptedBody, &obj); err != nil {
		return fmt.Errorf("unable to unmarshal json : %s", err.Error())
	}

	return nil
}
