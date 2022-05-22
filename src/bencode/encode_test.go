package bencode

import (
	"bytes"
	"fmt"
	"testing"
)

func TestEncodeInvalidType(t *testing.T) {
	_, err := Encode(true)
	if err == nil {
		t.Error("expected failure when encoding boolean")
	}
}

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

	// Should fail
	_, err := Encode("")
	if err == nil {
		t.Error("Should have panicked on empty string")
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

	for i := 0; i < len(ints); i++ {
		str := fmt.Sprintf("i%ve", ints[i])
		bstr := []byte(str)

		r, err := Encode(ints[i])
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(bstr, r) {
			t.Errorf("Expected [%v]\nGot [%v]", str, string(r))
		}
	}
}

func TestEncodeDict(t *testing.T) {
	correct := []byte("d4:key07:a value4:key113:another value5:key10i43242e5:key11i-123e5:key128:value 124:key244:a longer value of a long length that is long4:key31:a4:key4i0e4:key5i-174e4:key6i63847e4:key7i23985623467e4:key8i238497823e4:key9i4723985ee")

	dict := Dict{
		"key0":  "a value",
		"key1":  "another value",
		"key2":  "a longer value of a long length that is long",
		"key3":  "a",
		"key4":  0,
		"key5":  -174,
		"key6":  63847,
		"key7":  23985623467,
		"key8":  238497823,
		"key9":  4723985,
		"key10": int32(43242),
		"key11": int16(-123),
		"key12": "value 12",
	}

	r, err := Encode(dict)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(r, correct) {
		t.Errorf("\nExpected [%v]\n     Got [%v]", string(correct), string(r))
	}

	// Encode bad dictionary
	badDicts := []Dict{
		{"key": true},
		{"key": ""},
		{"": "value"},
	}

	for _, d := range badDicts {
		_, err = Encode(d)
		if err == nil {
			t.Errorf("expected failure encoding dictionary [%v]", d)
		}
	}
}

func TestEncodeList(t *testing.T) {
	correct := []byte("l5:elem12:e217:the third elementi234ei4268376598235ei-2394785ei0ei2039759823755235ee")
	l := List{
		"elem1",
		"e2",
		"the third element",
		234,
		4268376598235,
		-2394785,
		0,
		2039759823755235,
	}

	r, err := Encode(l)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(r, correct) {
		t.Errorf("\nExpected [%v]\n     Got [%v]", string(correct), string(r))
	}
}
