package ibc

import (
	"testing"
)

var minorErrorStringMap = map[int]string{
	512:  "Fan Pressure",
	4096: "Reversed Flow",
}

func TestGetMinorErrorString(t *testing.T) {

	for k, v := range minorErrorStringMap {
		r := GetErrorString(k, 0, 0)
		if v != r {
			t.Errorf("GetErrorString is incorrect, got: %s, want: %s.", r, v)
		}
	}
}
