package data_provider_conversion

import (
	"bytes"
	"encoding/gob"
)

func SerializeGOB[ModelT any](model ModelT) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(model); err != nil {
		return nil, err
	} else {
		return &buf, nil
	}
}

func DeserializeGOB[ModelT any](bufString *string) (ModelT, error) {
	var model ModelT
	buf := bytes.NewBufferString(*bufString)
	if err := gob.NewDecoder(buf).Decode(&model); err != nil {
		return model, nil
	} else {
		return model, nil
	}
}
