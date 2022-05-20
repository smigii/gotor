package bencode

import (
	"errors"
	"fmt"
	"sort"
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

	ec.encode(data)

	return ec.data[0:len(ec.data)], err
}

func (ec *encoder) encode(data interface{}) {
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
		ec.encodeInt(data.(int64))
	case Dict:
		ec.encodeDict(data.(Dict))
	case List:
		ec.encodeList(data.(List))
	default:
		panic("unknown type")
	}
}

func (ec *encoder) encodeString(str string) {
	encStr := fmt.Sprintf("%v:%v", len(str), str)
	ec.write([]byte(encStr))
}

func (ec *encoder) encodeInt(i int64) {
	encStr := []byte(fmt.Sprintf("i%ve", i))
	ec.write(encStr)
}

func (ec *encoder) encodeDict(dict Dict) {
	ec.write([]byte{'d'})

	keys := make([]string, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}
	// Keys must be sorted as per bencode spec
	sort.Strings(keys)

	for _, key := range keys {
		ec.encodeString(key)
		ec.encode(dict[key])
	}

	ec.write([]byte{'e'})
}

func (ec *encoder) encodeList(list List) {
	ec.write([]byte{'l'})

	for _, val := range list {
		ec.encode(val)
	}

	ec.write([]byte{'e'})
}

type encoder struct {
	data []byte
}

func newEncoder() *encoder {
	return &encoder{
		data: make([]byte, 0, 16),
	}
}

func (ec *encoder) write(bstr []byte) {
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
		newSlice = append(newSlice, ec.data...)
		ec.data = newSlice
	}

	ec.data = append(ec.data, bstr...)
}
