package lib

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// ByteString is a byte array that serializes to hex
type ByteString []byte

// MarshalJSON serializes ByteArray to hex
func (s ByteString) MarshalJSON() ([]byte, error) {
	bytes, err := json.Marshal(fmt.Sprintf("%x", string(s)))
	return bytes, err
}

// UnmarshalJSON deserializes ByteArray to hex
func (s *ByteString) UnmarshalJSON(data []byte) error {
	var x string
	err := json.Unmarshal(data, &x)
	if err == nil {
		str, e := hex.DecodeString(x)
		*s = ByteString([]byte(str))
		err = e
	}

	return err
}

func (s *ByteString) Value() (driver.Value, error) {
	return []byte(*s), nil
}

func (s *ByteString) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	*s = b
	return nil
}
