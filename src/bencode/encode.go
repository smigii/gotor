package bencode

import (
	"fmt"
	"sort"
)

// ============================================================================
// Error ======================================================================

type EncoderError struct{ msg string }

func (ee *EncoderError) Error() string {
	return "encoder error: " + ee.msg
}

// ============================================================================
// Public =====================================================================

func Encode(data interface{}) (ret []byte, err error) {
	ec := newEncoder()

	defer func() {
		if r := recover(); r != nil {
			err = &EncoderError{fmt.Sprintf("caught panic [%v]", r)}
			ret = nil
		}
	}()

	err = ec.encode(data)
	return ec.result(), err
}

// ============================================================================
// Private ====================================================================

func (ec *encoder) encode(data interface{}) error {
	var err error
	switch data.(type) {
	case string:
		err = ec.encodeString(data.(string))
	case int:
		ec.encodeInt(int64(data.(int)))
	case int8:
		ec.encodeInt(int64(data.(int8)))
	case int16:
		ec.encodeInt(int64(data.(int16)))
	case int32:
		ec.encodeInt(int64(data.(int32)))
	case int64:
		ec.encodeInt(data.(int64))
	case Dict:
		err = ec.encodeDict(data.(Dict))
	case List:
		err = ec.encodeList(data.(List))
	default:
		return &EncoderError{fmt.Sprintf("unsupported bencoding type [%T]", data)}
	}
	return err
}

func (ec *encoder) encodeString(str string) error {
	if len(str) == 0 {
		return &EncoderError{"cannot encode an empty string"}
	}
	encStr := fmt.Sprintf("%v:%v", len(str), str)
	ec.write([]byte(encStr))
	return nil
}

func (ec *encoder) encodeInt(i int64) {
	encStr := []byte(fmt.Sprintf("i%ve", i))
	ec.write(encStr)
}

func (ec *encoder) encodeDict(dict Dict) error {
	ec.write([]byte{'d'})

	keys := make([]string, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}

	// Keys must be sorted as per bencode spec
	sort.Strings(keys)

	for _, key := range keys {
		err := ec.encodeString(key)
		if err != nil {
			return err
		}
		err = ec.encode(dict[key])
		if err != nil {
			return err
		}
	}

	ec.write([]byte{'e'})
	return nil
}

func (ec *encoder) encodeList(list List) error {
	ec.write([]byte{'l'})

	for _, val := range list {
		err := ec.encode(val)
		if err != nil {
			return err
		}
	}

	ec.write([]byte{'e'})
	return nil
}

// ============================================================================
// Encoder Struct =============================================================

type encoder struct {
	data []byte
}

func newEncoder() *encoder {
	return &encoder{
		data: make([]byte, 0, 32),
	}
}

func (ec *encoder) write(bstr []byte) {
	bstrLen := uint64(len(bstr))
	newLen := uint64(len(ec.data)) + bstrLen
	curCap := uint64(cap(ec.data))

	// Check if we need to resize
	if newLen >= curCap {
		newCap := curCap << 1
		for {
			if newCap > newLen {
				break
			}
			newCap <<= 1
		}

		newSlice := make([]byte, 0, newCap)
		newSlice = append(newSlice, ec.data...)
		ec.data = newSlice
	}

	ec.data = append(ec.data, bstr...)
}

func (ec *encoder) result() []byte {
	return ec.data[0:len(ec.data)]
}
