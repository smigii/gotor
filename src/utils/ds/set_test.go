package ds

import "testing"

type keyMock struct {
	key string
}

func (km *keyMock) Key() string {
	return km.key
}

func TestSet_Add(t *testing.T) {
	tests := []struct {
		name  string
		words []string
	}{
		{"", []string{"a"}},
		{"", []string{"a", "b"}},
		{"", []string{"a", "b", "c"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := MakeSet[*keyMock]()
			kms := make([]*keyMock, 0, len(tt.words))

			for i, word := range tt.words {
				km := &keyMock{key: word}
				kms = append(kms, km)
				s.Add(km)

				if s.Size() != i+1 {
					t.Errorf("Size expected %v, got %v", i+1, s.Size())
				}
			}

			for _, km := range kms {
				if !s.Has(km) {
					t.Errorf("Set could not find %v", km.Key())
				}
			}

		})
	}
}
