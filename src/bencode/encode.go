package bencode

import (
	"bytes"
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
	buf := bytes.Buffer{}

	defer func() {
		if r := recover(); r != nil {
			err = &EncoderError{fmt.Sprintf("caught panic [%v]", r)}
			ret = nil
		}
	}()

	err = encode(data, &buf)
	return buf.Bytes(), err
}

// ============================================================================
// Private ====================================================================

func encode(data interface{}, buf *bytes.Buffer) error {
	var err error

	switch data.(type) {
	case string:
		err = encodeString(data.(string), buf)
	case int:
		encodeInt(int64(data.(int)), buf)
	case int8:
		encodeInt(int64(data.(int8)), buf)
	case int16:
		encodeInt(int64(data.(int16)), buf)
	case int32:
		encodeInt(int64(data.(int32)), buf)
	case int64:
		encodeInt(data.(int64), buf)
	case Dict:
		err = encodeDict(data.(Dict), buf)
	case List:
		err = encodeList(data.(List), buf)
	default:
		return &EncoderError{fmt.Sprintf("unsupported bencoding type [%T]", data)}
	}
	return err
}

func encodeString(str string, buf *bytes.Buffer) error {
	if len(str) == 0 {
		return &EncoderError{"cannot encode an empty string"}
	}
	encStr := fmt.Sprintf("%v:%v", len(str), str)
	buf.WriteString(encStr)
	return nil
}

func encodeInt(i int64, buf *bytes.Buffer) {
	encStr := fmt.Sprintf("i%ve", i)
	buf.WriteString(encStr)
}

func encodeDict(dict Dict, buf *bytes.Buffer) error {
	buf.WriteByte('d')

	keys := make([]string, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}

	// Keys must be sorted as per bencode spec
	sort.Strings(keys)

	for _, key := range keys {
		err := encodeString(key, buf)
		if err != nil {
			return err
		}
		err = encode(dict[key], buf)
		if err != nil {
			return err
		}
	}

	buf.WriteByte('e')
	return nil
}

func encodeList(list List, buf *bytes.Buffer) error {
	buf.WriteByte('l')

	for _, val := range list {
		err := encode(val, buf)
		if err != nil {
			return err
		}
	}

	buf.WriteByte('e')
	return nil
}
