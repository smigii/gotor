package main

import "testing"

func Test_execute(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"t1", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := execute(); got != tt.want {
				t.Errorf("execute() = %v, want %v", got, tt.want)
			}
		})
	}
}
