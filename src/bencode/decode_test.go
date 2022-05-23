package bencode

import (
	"fmt"
	"testing"
)

// https://adrianstoll.com/bencoding/

func TestDecode(t *testing.T) {

	t1 := []byte("d5:hello5:worlde")

	d, err := Decode(t1)
	if err != nil {
		t.Error(err)
	}
	dict := d.(Dict)

	if dict["hello"] != "world" {
		t.Errorf("Expected [world], got [%v]", dict["hello"])
	}

}

func TestDecodeString(t *testing.T) {
	// Test good strings
	words := []string{
		"a",
		"a word",
		"a fairly long word",
		"ahksudh akwughd uagh dfoih dfiaslhd iowlh fdoihdiash dkahfl halidhfiashialsfahsilfh ailsd hilshflas fklklakfdlkasfh",
	}

	var str string
	var bstr []byte

	for i := 0; i < len(words); i++ {

		str = fmt.Sprintf("%v:%v", len(words[i]), words[i])
		bstr = []byte(str)
		r, err := Decode(bstr)

		if err != nil {
			t.Error(err)
		}

		res, ok := r.(string)
		if !ok {
			t.Errorf("Failed converting to string: %v", r)
		}

		if res != words[i] {
			t.Errorf("Expected [%v], got [%v]", words[i], res)
		}
	}

	// Test bad strings
	badStrs := []string{
		"12:",
		"10:hi",
	}

	for i := 0; i < len(badStrs); i++ {
		bstr = []byte(badStrs[i])
		_, err := Decode(bstr)
		if err == nil {
			t.Errorf("Expected error on input [%v]", badStrs[i])
		}
	}
}

func TestDecodeInt(t *testing.T) {
	// Test good ints
	ints := []int64{
		0,
		1,
		-1,
		6,
		38,
		435,
		34635254,
		-2000000000,
		2000000001,
	}

	var str string
	var bstr []byte

	for i := 0; i < len(ints); i++ {

		str = fmt.Sprintf("i%ve", ints[i])
		bstr = []byte(str)
		r, err := Decode(bstr)

		if err != nil {
			t.Error(err)
		}

		res, ok := r.(int64)
		if !ok {
			t.Errorf("Failed converting to int: %v", r)
		}

		if res != ints[i] {
			t.Errorf("Expected [%v], got [%v]", ints[i], res)
		}
	}

	// Test bad ints
	badStrs := []string{
		"ie",
		"i00e",
		"ig30e",
		"i1o1e",
		"i43",
	}

	for i := 0; i < len(badStrs); i++ {

		bstr = []byte(badStrs[i])
		_, err := Decode(bstr)

		if err == nil {
			t.Errorf("Expected error on input [%v]", badStrs[i])
		}
	}

}

func TestDecodeDict(t *testing.T) {

	/*
		{
		    "key1": "value1",
		    "key2": "value2",
		    "key3": 3,
		    "key4": 4,
		    "dict2": {
		        "key5": 5,
		        "key6": 6,
		        "dict3": {
		            "key7": 7
		        }
		    },
		    "dict4": {
		        "key8": 8
		    }
		}
	*/
	bytes := []byte("d5:dict2d5:dict3d4:key7i7ee4:key5i5e4:key6i6ee5:dict4d4:key8i8ee4:key16:value14:key26:value24:key3i3e4:key4i4ee")
	d, err := Decode(bytes)

	if err != nil {
		t.Error(err)
	}

	dict1, ok := d.(Dict)
	checkOk(ok, "failed type conversion d1", t)
	dict2, ok := dict1["dict2"].(Dict)
	checkOk(ok, "failed type conversion d2", t)
	dict3, ok := dict2["dict3"].(Dict)
	checkOk(ok, "failed type conversion d3", t)
	dict4, ok := dict1["dict4"].(Dict)
	checkOk(ok, "failed type conversion d4", t)

	checkKey(dict1, "key1", "value1", t)
	checkKey(dict1, "key2", "value2", t)
	checkKey(dict1, "key3", int64(3), t)
	checkKey(dict1, "key4", int64(4), t)
	checkKey(dict2, "key5", int64(5), t)
	checkKey(dict2, "key6", int64(6), t)
	checkKey(dict3, "key7", int64(7), t)
	checkKey(dict4, "key8", int64(8), t)
}

func checkOk(ok bool, msg string, t *testing.T) {
	if !ok {
		t.Error(msg)
	}
}

func checkKey(dict Dict, key string, val interface{}, t *testing.T) {
	if dict[key] != val {
		t.Errorf("Expected [%v] got [%v]", val, dict[key])
	}
}

func TestDecodeList(t *testing.T) {

	/*
		[
		    "e1",
		    "e2",
		    "e3",
		    4,
		    5,
		    6,
		    [
		        "e7"
		    ],
		    [
		        "e8",
		        [
		            "e9"
		        ]
		    ]
		]
	*/
	bytes := []byte("l2:e12:e22:e3i4ei5ei6el2:e7el2:e8l2:e9eee")
	l, err := Decode(bytes)

	if err != nil {
		t.Error(err)
	}

	list1, ok := l.(List)
	checkOk(ok, "Failed converting list 1", t)
	list2, ok := list1[6].(List)
	checkOk(ok, "Failed converting list 2", t)
	list3, ok := list1[7].(List)
	checkOk(ok, "Failed converting list 3", t)
	list4, ok := list3[1].(List)
	checkOk(ok, "Failed converting list 4", t)

	checkVal(list1, 0, "e1", t)
	checkVal(list1, 5, int64(6), t)
	checkVal(list2, 0, "e7", t)
	checkVal(list3, 0, "e8", t)
	checkVal(list4, 0, "e9", t)
}

func checkVal(list List, idx int, val interface{}, t *testing.T) {
	if idx >= len(list) {
		t.Errorf("%v out of range for list -- %v", idx, list)
	}

	if list[idx] != val {
		t.Errorf("Expected [%v] got [%v]", val, list[idx])
	}
}
