package bencode

import (
	"strconv"
)

// ============================================================================
// Error ======================================================================

type DecoderError struct{ msg string }

func (de *DecoderError) Error() string {
	return "decoder error: " + de.msg
}

// ============================================================================
// Public =====================================================================

func Decode(data []byte) (ret interface{}, err error) {

	dc := decoder{data: data, curs: 0}

	defer func() {
		r := recover()
		if r != nil {
			if dc.curs >= uint64(len(dc.data)) {
				err = &DecoderError{"read end of data before completion"}
			} else {
				err = &DecoderError{"caught panic"}
			}
			ret = nil
		}
	}()

	firstByte := dc.curByte()

	switch firstByte {
	case 'd':
		ret, err = dc.decodeDict()
	case 'l':
		ret, err = dc.decodeList()
	case 'i':
		ret, err = dc.decodeInt()
	default:
		ret, err = dc.decodeString()
	}

	return ret, err
}

// ============================================================================
// Private ====================================================================

func (dc *decoder) decodeDict() (Dict, error) {

	if dc.curByte() != 'd' {
		return nil, &DecoderError{"not pointing to dict"}
	}

	dc.curs++
	dict := make(map[string]interface{})

	for {
		if dc.curByte() == 'e' {
			dc.curs++
			break
		}

		key, err := dc.decodeString()
		if err != nil {
			return nil, err
		}

		var val interface{}

		switch dc.curByte() {
		case 'i':
			val, err = dc.decodeInt()
		case 'l':
			val, err = dc.decodeList()
		case 'd':
			val, err = dc.decodeDict()
		default: // another string
			val, err = dc.decodeString()
		}

		if err != nil {
			return nil, err
		}
		dict[key] = val
	}

	return dict, nil
}

func (dc *decoder) decodeList() (List, error) {
	if dc.curByte() != 'l' {
		return nil, &DecoderError{"not pointing to dict"}
	}

	dc.curs++
	var list List

	for {
		if dc.curByte() == 'e' {
			dc.curs++
			break
		}

		var err error
		var val interface{}

		switch dc.curByte() {
		case 'i': // 'i' integer
			val, err = dc.decodeInt()
		case 'l': // 'l' list
			val, err = dc.decodeList()
		case 'd': // 'd' dict
			val, err = dc.decodeDict()
		default: // another string
			val, err = dc.decodeString()
		}

		if err != nil {
			return nil, err
		}
		list = append(list, val)
	}

	return list, nil
}

func (dc *decoder) decodeString() (string, error) {
	var strLen string

	for {
		if dc.curByte() >= '0' && dc.curByte() <= '9' {
			strLen += string(dc.curByte())
			dc.curs++
		} else if dc.curByte() == ':' {
			dc.curs++
			break
		} else {
			return "", &DecoderError{"bad string at byte " + strconv.FormatUint(dc.curs, 10)}
		}
	}

	lenVal, err := strconv.Atoi(strLen)
	if err != nil {
		return "", err
	}

	str := string(dc.data[dc.curs : dc.curs+uint64(lenVal)])
	dc.curs += uint64(lenVal)

	return str, nil
}

func (dc *decoder) decodeInt() (int64, error) {
	// 1. Read 'i'
	// 2. Build string until 'e'

	if dc.curByte() != 'i' {
		return 0, &DecoderError{"invalid call to decodeInt, not pointing at 'i'"}
	}

	dc.curs++
	var strInt string
	negMult := int64(1)

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
			return 0, &DecoderError{"bad int (leading 0s) at byte " + strconv.FormatUint(dc.curs, 10)}
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
			return 0, &DecoderError{"bad int at byte " + strconv.FormatUint(dc.curs, 10)}
		}
	}

	val, err := strconv.ParseInt(strInt, 10, 64)
	if err != nil {
		return 0, err
	}

	return val * negMult, nil
}

// ============================================================================
// Decoder Struct =============================================================

type decoder struct {
	data []byte
	curs uint64
}

func (dc *decoder) curByte() byte {
	return dc.data[dc.curs]
}

func (dc *decoder) nextByte() byte {
	return dc.data[dc.curs+1]
}
