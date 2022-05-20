package bencode

import (
	"bytes"
	"fmt"
	"testing"
)

func TestEncodeString(t *testing.T) {
	strs := []string{
		"a",
		"abc",
		"hello there person how are you",
	}

	// Test enc(dec(x)) == x
	for i := 0; i < len(strs); i++ {
		str := fmt.Sprintf("%v:%v", len(strs[i]), strs[i])
		bstr := []byte(str)

		r, err := Encode(strs[i])
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(bstr, r) {
			t.Errorf("Expected [%v]\nGot [%v]", str, string(r))
		}
	}
}

func TestEncodeInt(t *testing.T) {

	ints := []interface{}{
		0,
		int(-23),
		int(135423),
		int8(45),
		int8(-23),
		int16(432),
		int16(-213),
		int32(324),
		int32(-6435),
		int64(-34543234),
		int64(42346785983),
	}

	// Test enc(dec(x)) == x
	for i := 0; i < len(ints); i++ {
		str := fmt.Sprintf("i%ve", ints[i])
		bstr := []byte(str)
		//fmt.Println(str)

		r, err := Encode(ints[i])
		if err != nil {
			t.Error(err)
		}

		//fmt.Println(string(r))
		if !bytes.Equal(bstr, r) {
			t.Errorf("Expected [%v]\nGot [%v]", str, string(r))
		}
	}
}

func TestEncodeDict(t *testing.T) {

	d := make(Dict)
	d["key1"] = "value1"
	d["key2"] = "value2"
	r, err := Encode(d)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(r))

}

func TestEncodeList(t *testing.T) {

	l := make(List, 0, 32)
	l = append(l, "hello")
	l = append(l, "another")
	l = append(l, "word")
	l = append(l, "asd")
	l = append(l, 123)
	l = append(l, -123)
	l = append(l, 0)
	l = append(l, 5734987)

	r, err := Encode(l)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(string(r))
}
