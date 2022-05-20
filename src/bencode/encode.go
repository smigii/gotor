package bencode

import (
	"errors"
	"fmt"
)

func Encode(data interface{}) (ret []byte, err error) {

	//ec := encoder{data: data}
	err = nil

	defer func() {
		r := recover()
		if r != nil {
			err = errors.New("caught panic")
			ret = nil
		}
	}()

	switch data.(type) {
	case string:
		ret = encodeString(data.(string))
	case int:
		ret = encodeInt(int64(data.(int)))
	case int8:
		ret = encodeInt(int64(data.(int8)))
	case int16:
		ret = encodeInt(int64(data.(int16)))
	case int32:
		ret = encodeInt(int64(data.(int32)))
	case int64:
		ret = encodeInt(int64(data.(int64)))
	default:
		ret = []byte{}
		err = errors.New("unknown type")
	}

	return ret, err

}

func encodeString(str string) []byte {
	return []byte(fmt.Sprintf("%v:%v", len(str), str))
}

func encodeInt(i int64) []byte {
	return []byte(fmt.Sprintf("i%ve", i))
}

type encoder struct {
	data interface{}
}
