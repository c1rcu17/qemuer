//go:generate go run gen/gen.go

package static

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
)

func OpenResource(name string) (io.ReadCloser, error) {
	if b64, exists := resources[name]; !exists {
		return nil, fmt.Errorf("resource not found %s", name)
	} else {
		if compressed, err := base64.StdEncoding.DecodeString(b64); err != nil {
			return nil, err
		} else {
			if reader, err := gzip.NewReader(bytes.NewReader(compressed)); err != nil {
				return nil, err
			} else {
				return reader, nil
			}
		}
	}
}
