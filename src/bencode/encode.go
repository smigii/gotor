package bencode

import (
	"errors"
	"fmt"
)

func Encode(data interface{}) (ret []byte, err error) {

	ec := newEncoder()
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
		ec.encodeString(data.(string))
	case int:
		ec.encodeInt(int64(data.(int)))
	case int8:
		ec.encodeInt(int64(data.(int8)))
	case int16:
		ec.encodeInt(int64(data.(int16)))
	case int32:
		ec.encodeInt(int64(data.(int32)))
	case int64:
		ec.encodeInt(int64(data.(int64)))
	default:
		ret = []byte{}
		err = errors.New("unknown type")
	}

	return ec.data[0:len(ec.data)], err
}

func (ec *encoder) encodeString(str string) {
	encStr := fmt.Sprintf("%v:%v", len(str), str)
	ec.copy([]byte(encStr))
}

func (ec *encoder) encodeInt(i int64) {
	encStr := []byte(fmt.Sprintf("i%ve", i))
	ec.copy(encStr)
}

func encodeDict(dict Dict) []byte {
	keys := make([]interface{}, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}
	fmt.Println(keys)

	return []byte{}
}

type encoder struct {
	data []byte
}

func newEncoder() *encoder {
	return &encoder{
		data: make([]byte, 0, 16),
	}
}

func (ec *encoder) copy(bstr []byte) {
	bstrLen := uint64(len(bstr))
	newLen := uint64(len(ec.data)) + bstrLen
	curCap := uint64(cap(ec.data))

	if newLen >= curCap {
		newCap := curCap << 1
		for {
			if newCap > newLen {
				break
			}
			newCap <<= 1
		}

		newSlice := make([]byte, 0, newCap)
		copy(newSlice, ec.data)
		ec.data = newSlice
	}

	ec.data = append(ec.data, bstr...)
}
