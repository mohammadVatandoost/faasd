package lru

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func getRealSizeOf(v interface{}) (int, error) {
	b := new(bytes.Buffer)
	if err := gob.NewEncoder(b).Encode(v); err != nil {
		fmt.Println("***** can not getRealSizeOf *****")
		return 0, err
	}
	return b.Len(), nil
}

