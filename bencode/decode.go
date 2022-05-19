package bencode

import (
	"errors"
	"fmt"
	"strconv"
)

const decodeErrMsg string = "decode error: "

type dataCursor struct {
	data []byte
	curs uint64
}

func (dc *dataCursor) curByte() byte {
	return dc.data[dc.curs]
}

func (dc *dataCursor) nextByte() byte {
	return dc.data[dc.curs+1]
}

func Decode(data []byte) (interface{}, error) {

	dc := dataCursor{data: data, curs: 0}
	firstByte := dc.curByte()

	switch firstByte {
	case 'd':
		d, err := decodeDict(&dc)
		return d, err
	case 'l':
		l, err := decodeList(&dc)
		return l, err
	default:
		msg := fmt.Sprintf("invalid first byte [%v]", firstByte)
		return nil, errors.New(msg)
	}

}

func decodeDict(dc *dataCursor) (map[string]interface{}, error) {

	if dc.curByte() != 'd' {
		return nil, errors.New(decodeErrMsg + "not pointing to dict")
	}

	dc.curs++
	dict := make(map[string]interface{})

	for {
		if dc.curByte() == 'e' {
			dc.curs++
			break
		}

		key, err := decodeString(dc)
		if err != nil {
			panic(err)
		}

		var val interface{}

		switch dc.curByte() {
		case 'i': // 'i' integer
			val, err = decodeInt(dc)
		case 'l': // 'l' list
			val, err = decodeList(dc)
		case 'd': // 'd' dict
			val, err = decodeDict(dc)
		default: // another string
			val, err = decodeString(dc)
		}

		if err != nil {
			panic(err)
		}
		dict[key] = val
	}

	return dict, nil
}

func decodeList(dc *dataCursor) ([]interface{}, error) {
	if dc.curByte() != 'l' {
		return nil, errors.New(decodeErrMsg + "not pointing to dict")
	}

	dc.curs++
	var list []interface{}

	for {
		if dc.curByte() == 'e' {
			dc.curs++
			break
		}

		var err error
		var val interface{}

		switch dc.curByte() {
		case 'i': // 'i' integer
			val, err = decodeInt(dc)
		case 'l': // 'l' list
			val, err = decodeList(dc)
		case 'd': // 'd' dict
			val, err = decodeDict(dc)
		default: // another string
			val, err = decodeString(dc)
		}

		if err != nil {
			panic(err)
		}
		list = append(list, val)
	}

	return list, nil
}

func decodeString(dc *dataCursor) (string, error) {
	// 1. Read up to ':' for length n
	// 2. Read n bytes for string

	//var startIdx = dc.curs
	var strLen string

	for {
		if dc.curByte() >= '0' && dc.curByte() <= '9' {
			strLen += string(dc.curByte())
			dc.curs++
		} else if dc.curByte() == ':' {
			dc.curs++
			break
		} else {
			return "", errors.New(decodeErrMsg + "bad string at byte " + strconv.FormatUint(dc.curs, 10))
		}
	}

	lenVal, err := strconv.Atoi(strLen)
	if err != nil {
		panic(err)
	}

	str := string(dc.data[dc.curs : dc.curs+uint64(lenVal)])
	dc.curs += uint64(lenVal)

	return str, nil
}

func decodeInt(dc *dataCursor) (int, error) {
	// 1. Read 'i'
	// 2. Build string until 'e'

	if dc.curByte() != 'i' {
		return 0, errors.New("invalid call to decodeInt, not pointing at 'i'")
	}

	dc.curs++
	var strInt string
	negMult := 1

	// Handle negative ints
	if dc.curByte() == '-' {
		negMult = -1
		dc.curs++
	}

	// Should not have leading 0s, unless value is 0
	if dc.curByte() == '0' {
		strInt += string(dc.curByte())
		dc.curs++
		if dc.curByte() != 'e' {
			return 0, errors.New(decodeErrMsg + "bad int (leading 0s) at byte " + strconv.FormatUint(dc.curs, 10))
		}
	}

	// Read remaining bytes until 'e'
	for {
		if dc.curByte() >= '0' && dc.curByte() <= '9' {
			strInt += string(dc.curByte())
			dc.curs++
		} else if dc.curByte() == 'e' {
			dc.curs++
			break
		} else {
			return 0, errors.New(decodeErrMsg + "bad int at byte " + strconv.FormatUint(dc.curs, 10))
		}
	}

	val, err := strconv.Atoi(strInt)
	if err != nil {
		panic(err)
	}

	return val * negMult, nil
}
