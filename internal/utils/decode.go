package utils

import (
	"encoding/json"
	"io"
)

func UnmarshalReqBody(body io.ReadCloser, output interface{}) (err error) {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(output)
	return err
}

func UnmarshalReqBodyAllowUnknown(body io.ReadCloser, output interface{}) (err error) {
	decoder := json.NewDecoder(body)
	err = decoder.Decode(output)
	return err
}

func DecodeReqBodyAsString(body io.Reader) (string, error) {
	resp, err := io.ReadAll(body)
	return string(resp), err
}
