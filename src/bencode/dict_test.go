package bencode

import (
	"reflect"
	"testing"
)

// Dummy globals
var subDict Dict
var subList List

func testDict() Dict {
	subDict = make(Dict)
	d := make(Dict)
	d["str"] = "a string"
	d["int"] = int64(-324)
	d["uint"] = int64(4273)
	d["dict"] = subDict
	d["list"] = subList
	return d
}

func TestDict_GetString(t *testing.T) {
	tests := []struct {
		name    string
		d       Dict
		key     string
		want    string
		wantErr bool
	}{
		{"good key", testDict(), "str", "a string", false},
		{"bad type", testDict(), "int", "", true},
		{"bad key", testDict(), "bad key", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.GetString(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDict_GetInt(t *testing.T) {
	tests := []struct {
		name    string
		d       Dict
		key     string
		want    int64
		wantErr bool
	}{
		{"good key", testDict(), "int", -324, false},
		{"bad type", testDict(), "str", 0, true},
		{"bad key", testDict(), "bad key", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.GetInt(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetInt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDict_GetUint(t *testing.T) {
	tests := []struct {
		name    string
		d       Dict
		key     string
		want    uint64
		wantErr bool
	}{
		{"good key", testDict(), "uint", 4273, false},
		{"bad type", testDict(), "str", 0, true},
		{"bad key", testDict(), "bad key", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.GetUint(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetUint() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDict_GetDict(t *testing.T) {
	tests := []struct {
		name    string
		d       Dict
		key     string
		want    Dict
		wantErr bool
	}{
		{"good key", testDict(), "dict", subDict, false},
		{"bad type", testDict(), "str", nil, true},
		{"bad key", testDict(), "bad key", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.GetDict(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDict() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDict() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDict_GetList(t *testing.T) {
	tests := []struct {
		name    string
		d       Dict
		key     string
		want    List
		wantErr bool
	}{
		{"good key", testDict(), "list", subList, false},
		{"bad type", testDict(), "str", nil, true},
		{"bad key", testDict(), "bad key", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.GetList(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetList() got = %v, want %v", got, tt.want)
			}
		})
	}
}
