package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"net/http"

	"github.com/twmb/murmur3"
)

func GetIconHashFromURL(iconURL string) (string, error) {
	resp, err := http.Get(iconURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return Mmh3Hash32(standardBase64(data)), nil
}

func Mmh3Hash32(data []byte) string {
	var h32 hash.Hash32 = murmur3.New32()
	h32.Write(data)
	return fmt.Sprintf("%d", int32(h32.Sum32()))
}

func standardBase64(data []byte) []byte {
	encoded := base64.StdEncoding.EncodeToString(data)
	var buffer bytes.Buffer
	for i := 0; i < len(encoded); i++ {
		buffer.WriteByte(encoded[i])
		if (i+1)%76 == 0 {
			buffer.WriteByte('\n')
		}
	}
	buffer.WriteByte('\n')
	return buffer.Bytes()
}
