package utils

import (
	"encoding/json"
	"io"
)

func UnmarshalReqBody(body io.ReadCloser, output interface{}) (err error) {
	decoder := json.NewDecoder(body)
	err = decoder.Decode(output)
	return err
}
