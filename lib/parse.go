package lib

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/bitcoinschema/go-bitcoin"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
)

var PATTERN []byte
var MAP = "1PuQa7K62MiKCtssSLKy1kh56WWU7MtUR5"
var B = "19HxigV4QyBv3tHpQVcUEQyq1pzZVdoAut"

var OrdLockPrefix []byte
var OrdLockSuffix []byte

func init() {
	val, err := hex.DecodeString("0063036f7264")
	if err != nil {
		log.Panic(err)
	}
	PATTERN = val

	OrdLockPrefix, _ = hex.DecodeString("2097dfd76851bf465e8f715593b217714858bbe9570ff3bd5e33840a34e20ff0262102ba79df5f8ae7604a9830f03c7933028186aede0675a16f025dc4f8be8eec0382201008ce7480da41702918d1ec8e6849ba32b4d65b1e40dc669c31a1e6306b266c0000")
	OrdLockSuffix, _ = hex.DecodeString("615179547a75537a537a537a0079537a75527a527a7575615579008763567901c161517957795779210ac407f0e4bd44bfc207355a778b046225a7068fc59ee7eda43ad905aadbffc800206c266b30e6a1319c66dc401e5bd6b432ba49688eecd118297041da8074ce081059795679615679aa0079610079517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e01007e81517a75615779567956795679567961537956795479577995939521414136d08c5ed2bf3ba048afe6dcaebafeffffffffffffffffffffffffffffff00517951796151795179970079009f63007952799367007968517a75517a75517a7561527a75517a517951795296a0630079527994527a75517a6853798277527982775379012080517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e01205279947f7754537993527993013051797e527e54797e58797e527e53797e52797e57797e0079517a75517a75517a75517a75517a75517a75517a75517a75517a75517a75517a75517a75517a756100795779ac517a75517a75517a75517a75517a75517a75517a75517a75517a7561517a75517a756169587951797e58797eaa577961007982775179517958947f7551790128947f77517a75517a75618777777777777777777767557951876351795779a9876957795779ac777777777777777767006868")
}

type OpPart struct {
	OpCode byte
	Data   []byte
	Len    uint32
}

func ReadOp(b []byte, idx *int) (op *OpPart, err error) {
	if len(b) <= *idx {
		// log.Panicf("ReadOp: %d %d", len(b), *idx)
		err = fmt.Errorf("ReadOp: %d %d", len(b), *idx)
		return
	}
	switch b[*idx] {
	case bscript.OpPUSHDATA1:
		if len(b) < *idx+2 {
			err = bscript.ErrDataTooSmall
			return
		}

		l := int(b[*idx+1])
		*idx += 2

		if len(b) < *idx+l {
			err = bscript.ErrDataTooSmall
			return
		}

		op = &OpPart{OpCode: bscript.OpPUSHDATA1, Data: b[*idx : *idx+l]}
		*idx += l

	case bscript.OpPUSHDATA2:
		if len(b) < *idx+3 {
			err = bscript.ErrDataTooSmall
			return
		}

		l := int(binary.LittleEndian.Uint16(b[*idx+1:]))
		*idx += 3

		if len(b) < *idx+l {
			err = bscript.ErrDataTooSmall
			return
		}

		op = &OpPart{OpCode: bscript.OpPUSHDATA2, Data: b[*idx : *idx+l]}
		*idx += l

	case bscript.OpPUSHDATA4:
		if len(b) < *idx+5 {
			err = bscript.ErrDataTooSmall
			return
		}

		l := int(binary.LittleEndian.Uint32(b[*idx+1:]))
		*idx += 5

		if len(b) < *idx+l {
			err = bscript.ErrDataTooSmall
			return
		}

		op = &OpPart{OpCode: bscript.OpPUSHDATA4, Data: b[*idx : *idx+l]}
		*idx += l

	default:
		if b[*idx] >= 0x01 && b[*idx] < bscript.OpPUSHDATA1 {
			l := b[*idx]
			if len(b) < *idx+int(1+l) {
				err = bscript.ErrDataTooSmall
				return
			}
			op = &OpPart{OpCode: b[*idx], Data: b[*idx+1 : *idx+int(l+1)]}
			*idx += int(1 + l)
		} else {
			op = &OpPart{OpCode: b[*idx]}
			*idx++
		}
	}

	return
}

func ParseBitcom(txo *Txo, idx *int) (err error) {
	script := *txo.Tx.Outputs[txo.Vout].LockingScript
	tx := txo.Tx
	d := txo.Data

	startIdx := *idx
	op, err := ReadOp(script, idx)
	if err != nil {
		return
	}
	switch string(op.Data) {
	case MAP:
		op, err = ReadOp(script, idx)
		if err != nil {
			return
		}
		if string(op.Data) != "SET" {
			return nil
		}
		d.Map = map[string]interface{}{}
		for {
			prevIdx := *idx
			op, err = ReadOp(script, idx)
			if err != nil || op.OpCode == bscript.OpRETURN || (op.OpCode == 1 && op.Data[0] == '|') {
				*idx = prevIdx
				break
			}
			opKey := op.Data
			prevIdx = *idx
			op, err = ReadOp(script, idx)
			if err != nil || op.OpCode == bscript.OpRETURN || (op.OpCode == 1 && op.Data[0] == '|') {
				*idx = prevIdx
				break
			}

			if !utf8.Valid(opKey) || !utf8.Valid(op.Data) {
				continue
			}

			if len(opKey) == 1 && opKey[0] == 0 {
				opKey = []byte{}
			}
			if len(op.Data) == 1 && op.Data[0] == 0 {
				op.Data = []byte{}
			}
			d.Map[string(opKey)] = string(op.Data)
		}
		if val, ok := d.Map["subTypeData"]; ok {
			var subTypeData json.RawMessage
			if err := json.Unmarshal([]byte(val.(string)), &subTypeData); err == nil {
				d.Map["subTypeData"] = subTypeData
			}
		}
		return nil
	case B:
		d.B = &File{}
		for i := 0; i < 4; i++ {
			prevIdx := *idx
			op, err = ReadOp(script, idx)
			if err != nil || op.OpCode == bscript.OpRETURN || (op.OpCode == 1 && op.Data[0] == '|') {
				*idx = prevIdx
				break
			}

			switch i {
			case 0:
				d.B.Content = op.Data
			case 1:
				d.B.Type = string(op.Data)
			case 2:
				d.B.Encoding = string(op.Data)
			case 3:
				d.B.Name = string(op.Data)
			}
		}
		hash := sha256.Sum256(d.B.Content)
		d.B.Size = uint32(len(d.B.Content))
		d.B.Hash = hash[:]
	case "SIGMA":
		sigma := &Sigma{}
		for i := 0; i < 4; i++ {
			prevIdx := *idx
			op, err = ReadOp(script, idx)
			if err != nil || op.OpCode == bscript.OpRETURN || (op.OpCode == 1 && op.Data[0] == '|') {
				*idx = prevIdx
				break
			}

			switch i {
			case 0:
				sigma.Algorithm = string(op.Data)
			case 1:
				sigma.Address = string(op.Data)
			case 2:
				sigma.Signature = op.Data
			case 3:
				vin, err := strconv.ParseInt(string(op.Data), 10, 32)
				if err == nil {
					sigma.Vin = uint32(vin)
				}
			}
		}
		d.Sigmas = append(d.Sigmas, sigma)

		outpoint := tx.Inputs[sigma.Vin].PreviousTxID()
		outpoint = binary.LittleEndian.AppendUint32(outpoint, tx.Inputs[sigma.Vin].PreviousTxOutIndex)
		// fmt.Printf("outpoint %x\n", outpoint)
		inputHash := sha256.Sum256(outpoint)
		// fmt.Printf("ihash: %x\n", inputHash)
		var scriptBuf []byte
		if script[startIdx-1] == bscript.OpRETURN {
			scriptBuf = script[:startIdx-1]
		} else if script[startIdx-1] == '|' {
			scriptBuf = script[:startIdx-2]
		} else {
			return nil
		}
		// fmt.Printf("scriptBuf %x\n", scriptBuf)
		outputHash := sha256.Sum256(scriptBuf)
		// fmt.Printf("ohash: %x\n", outputHash)
		msgHash := sha256.Sum256(append(inputHash[:], outputHash[:]...))
		// fmt.Printf("msghash: %x\n", msgHash)
		err = bitcoin.VerifyMessage(sigma.Address,
			base64.StdEncoding.EncodeToString(sigma.Signature),
			string(msgHash[:]),
		)
		if err != nil {
			fmt.Println("Error verifying signature", err)
			return nil
		}
		sigma.Valid = true
	default:
		*idx--
	}
	return
}

func ParseScript(txo *Txo) {
	d := &TxoData{}
	txo.Data = d
	script := *txo.Tx.Outputs[txo.Vout].LockingScript

	start := 0
	if len(script) >= 25 && bscript.NewFromBytes(script[:25]).IsP2PKH() {
		txo.PKHash = []byte(script[3:23])
		start = 25
	}

	var opFalse int
	var opIf int
	var opReturn int
	for i := start; i < len(script); {
		startI := i
		op, err := ReadOp(script, &i)
		if err != nil {
			break
		}
		// fmt.Println(prevI, i, op)
		switch op.OpCode {
		case bscript.Op0:
			opFalse = startI
		case bscript.OpIF:
			opIf = startI
		case bscript.OpRETURN:
			if opReturn == 0 {
				opReturn = startI
			}
			err = ParseBitcom(txo, &i)
			if err != nil {
				log.Println("Error parsing bitcom", err)
				continue
			}
		case bscript.OpDATA1:
			if op.Data[0] == '|' && opReturn > 0 {
				err = ParseBitcom(txo, &i)
				if err != nil {
					log.Println("Error parsing bitcom", err)
					continue
				}
			}
		}

		if bytes.Equal(op.Data, []byte("ord")) && opIf == startI-1 && opFalse == startI-2 {
			ins := &Inscription{
				File: &File{},
			}
		ordLoop:
			for {
				op, err = ReadOp(script, &i)
				if err != nil {
					break
				}
				switch op.OpCode {
				case bscript.Op0:
					op, err = ReadOp(script, &i)
					if err != nil {
						break ordLoop
					}
					ins.File.Content = op.Data
				case bscript.Op1:
					// case bscript.OpDATA1:
					// 	if op.OpCode == bscript.OpDATA1 && op.Data[0] != 1 {
					// 		continue
					// 	}
					op, err = ReadOp(script, &i)
					if err != nil {
						break ordLoop
					}
					if utf8.Valid(op.Data) {
						if len(op.Data) <= 256 {
							ins.File.Type = string(op.Data)
						} else {
							ins.File.Type = string(op.Data[:256])
						}
					}
				case bscript.OpENDIF:
					break ordLoop
				}
			}
			ins.File.Size = uint32(len(ins.File.Content))
			hash := sha256.Sum256(ins.File.Content)
			ins.File.Hash = hash[:]
			d.Inscription = ins
			insType := "file"
			if ins.File.Size <= 1024 && utf8.Valid(ins.File.Content) && !bytes.Contains(ins.File.Content, []byte{0}) {
				mime := strings.ToLower(ins.File.Type)
				if strings.HasPrefix(mime, "application/bsv-20") ||
					strings.HasPrefix(mime, "text/plain") ||
					strings.HasPrefix(mime, "application/json") {

					var data json.RawMessage
					err = json.Unmarshal(ins.File.Content, &data)
					if err == nil {
						insType = "json"
						ins.Json = data
						if strings.HasPrefix(mime, "application/bsv-20") {
							d.Bsv20, _ = parseBsv20(ins.File, txo.Height)
						}
						if txo.Height != nil && *txo.Height < 793000 &&
							strings.HasPrefix(mime, "text/plain") {
							d.Bsv20, _ = parseBsv20(ins.File, txo.Height)
						}
						if d.Bsv20 != nil {
							txo.Data.Types = append(txo.Data.Types, "bsv20")
						}
					}
				}
				if strings.HasPrefix(mime, "text") {
					if insType == "file" {
						insType = "text"
					}
					ins.Text = string(ins.File.Content)
					re := regexp.MustCompile(`\W`)

					words := map[string]struct{}{}
					for _, word := range re.Split(ins.Text, -1) {
						if len(word) > 1 {
							words[word] = struct{}{}
						}
					}

					if len(words) > 1 {
						ins.Words = make([]string, 0, len(words))
						for word := range words {
							ins.Words = append(ins.Words, word)
						}
					}
				}
			}
			txo.Data.Types = append(txo.Data.Types, insType)
		}
	}

	ordLockPrefixIndex := bytes.Index(script, OrdLockPrefix)
	ordLockSuffixIndex := bytes.Index(script, OrdLockSuffix)
	if ordLockPrefixIndex > -1 && ordLockSuffixIndex > len(OrdLockPrefix) {
		ordLock := script[ordLockPrefixIndex+len(OrdLockPrefix) : ordLockSuffixIndex]
		if ordLockParts, err := bscript.DecodeParts(ordLock); err == nil {
			txo.PKHash = ordLockParts[0]
			payOutput := &bt.Output{}
			_, err = payOutput.ReadFrom(bytes.NewReader(ordLockParts[1]))
			if err == nil {
				d.Listing = &Listing{
					Price:  payOutput.Satoshis,
					PayOut: payOutput.Bytes(),
				}
			}
		}
	}
}
