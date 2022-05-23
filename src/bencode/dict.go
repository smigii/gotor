package bencode

import "fmt"

//type BenType interface {
//	int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64 | string | Dict | List
//}

type Dict map[string]interface{}
type List []interface{}

type DictBadTypeError struct {
	key   string
	tWant string
	tHave string
	dict  Dict
}

func (e *DictBadTypeError) Error() string {
	return fmt.Sprintf("bad type for key [%v] in dict [%p]\nrequest [%v]\n    got [%v]", e.key, e.dict, e.tWant, e.tHave)
}

type DictMissingKeyError struct {
	key  string
	dict Dict
}

func (e *DictMissingKeyError) Error() string {
	return fmt.Sprintf("missing key [%v] in bencode dictionary [%p]", e.key, &e.dict)
}

func (d Dict) get(key string) (interface{}, error) {
	if v, ok := d[key]; ok {
		return v, nil
	} else {
		return nil, &DictMissingKeyError{
			key:  key,
			dict: d,
		}
	}
}

// Not really a fan of this, even though it cuts down on
// code duplication, passing the dict as a parameter feels gross
//func Get[T BenType](dict Dict, key string) (T, error) {
//	var zero T
//
//	if val, ok := dict[key]; !ok {
//		return zero, &DictMissingKeyError{
//			key:  key,
//			dict: dict,
//		}
//	} else {
//		if ret, ok := val.(T); ok {
//			return ret, nil
//		} else {
//			return zero, &DictBadTypeError{
//				key:   key,
//				tWant: fmt.Sprintf("%T", zero),
//				tHave: fmt.Sprintf("%T", val),
//				dict:  dict,
//			}
//		}
//	}
//}

func (d Dict) GetString(key string) (string, error) {
	if val, e := d.get(key); e != nil {
		return "", e
	} else {
		if s, ok := val.(string); ok {
			return s, nil
		} else {
			return "", &DictBadTypeError{
				key:   key,
				tWant: "string",
				tHave: fmt.Sprintf("%T", val),
				dict:  d,
			}
		}
	}
}

func (d Dict) GetInt(key string) (int64, error) {
	if val, e := d.get(key); e != nil {
		return 0, e
	} else {
		if i, ok := val.(int64); ok {
			return i, nil
		} else {
			return 0, &DictBadTypeError{
				key:   key,
				tWant: "int64",
				tHave: fmt.Sprintf("%T", val),
				dict:  d,
			}
		}
	}
}

func (d Dict) GetUint(key string) (uint64, error) {
	v, err := d.GetInt(key)
	if err != nil {
		return 0, err
	}
	return uint64(v), nil
}

func (d Dict) GetDict(key string) (Dict, error) {
	if val, e := d.get(key); e != nil {
		return nil, e
	} else {
		if d, ok := val.(Dict); ok {
			return d, nil
		} else {
			return nil, &DictBadTypeError{
				key:   key,
				tWant: "dict",
				tHave: fmt.Sprintf("%T", val),
				dict:  d,
			}
		}
	}
}

func (d Dict) GetList(key string) (List, error) {
	if val, e := d.get(key); e != nil {
		return nil, e
	} else {
		if l, ok := val.(List); ok {
			return l, nil
		} else {
			return nil, &DictBadTypeError{
				key:   key,
				tWant: "list",
				tHave: fmt.Sprintf("%T", val),
				dict:  d,
			}
		}
	}
}
