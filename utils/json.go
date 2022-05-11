package utils

import (
	"bytes"
	"encoding/json"
	"io"
)

// ParseJsonDropKeys parses the given JSON input object to an object of given type, while omitting the specified keys on the way.
// This can be useful if parsing would normally fail due to ambiguous typing of some key, but that key is not of interest and can be dropped to avoid parse errors.
// Dropping keys only works on top level of the object.
func ParseJsonDropKeys[T any](r io.Reader, dropKeys ...string) (T, error) {
	var (
		result       T
		resultTmp    map[string]interface{}
		resultTmpBuf = new(bytes.Buffer)
	)
	if err := json.NewDecoder(r).Decode(&resultTmp); err != nil {
		return result, err
	}

	for _, k := range dropKeys {
		delete(resultTmp, k)
	}

	if err := json.NewEncoder(resultTmpBuf).Encode(resultTmp); err != nil {
		return result, err
	}
	if err := json.NewDecoder(resultTmpBuf).Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}
