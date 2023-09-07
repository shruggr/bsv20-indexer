package lib

import (
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type Outpoint []byte

func NewOutpoint(txid []byte, vout uint32) *Outpoint {
	o := Outpoint(binary.BigEndian.AppendUint32(txid, vout))
	return &o
}

func NewOutpointFromString(s string) (o *Outpoint, err error) {
	txid, err := hex.DecodeString(s[:64])
	if err != nil {
		return
	}
	vout, err := strconv.ParseUint(s[65:], 10, 32)
	if err != nil {
		return
	}
	origin := Outpoint(binary.BigEndian.AppendUint32(txid, uint32(vout)))
	o = &origin
	return
}

func (o *Outpoint) String() string {
	return fmt.Sprintf("%x_%d", (*o)[:32], binary.BigEndian.Uint32((*o)[32:]))
}

func (o *Outpoint) Txid() []byte {
	return (*o)[:32]
}

func (o *Outpoint) Vout() uint32 {
	return binary.BigEndian.Uint32((*o)[32:])
}

func (o Outpoint) MarshalJSON() (bytes []byte, err error) {
	if len(o) == 36 {
		bytes, err = json.Marshal(fmt.Sprintf("%x_%d", o[:32], binary.BigEndian.Uint32(o[32:])))
	}
	return bytes, err
}

// UnmarshalJSON deserializes Origin to string
func (o *Outpoint) UnmarshalJSON(data []byte) error {
	var x string
	err := json.Unmarshal(data, &x)
	if err == nil {
		txid, err := hex.DecodeString(x[:64])
		if err != nil {
			return err
		}
		vout, err := strconv.ParseUint(x[65:], 10, 32)
		if err != nil {
			return err
		}

		*o = Outpoint(binary.BigEndian.AppendUint32(txid, uint32(vout)))
	}

	return err
}

func (o Outpoint) Value() (driver.Value, error) {
	return []byte(o), nil
}

func (o *Outpoint) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	*o = b
	return nil
}
