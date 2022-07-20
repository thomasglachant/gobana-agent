package core

import (
	"embed"
	"fmt"
	"io"
)

var AssetFs embed.FS

func AssetAsReader(filename string) (io.Reader, error) {
	f, err := AssetFs.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read file : %s", err)
	}
	return f, nil
}
