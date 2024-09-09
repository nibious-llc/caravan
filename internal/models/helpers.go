package models

import (
	"encoding/json"
)

func MarshalObject(o any) ([]byte, error) {

	data, err := json.Marshal(o)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func UnmarshalObject(data []byte, v any) error {
	return json.Unmarshal(data, &v)
}
