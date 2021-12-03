package cache

import (
	"fmt"
	"testing"
)

func TestByteView(t *testing.T) {
	for _, s := range []string{"", "x", "yy"} {
		v := of([]byte(s))
		name := fmt.Sprintf("string %q, view %+v", s, v)
		if v.Len() != len(s) {
			t.Errorf("%s: Len = %d; want %d", name, v.Len(), len(s))
		}
		if v.String() != s {
			t.Errorf("%s: String = %q; want %q", name, v.String(), s)
		}
	}
}

// of returns a byte view of the []byte.
func of(x interface{}) ByteView {
	return ByteView{b: x.([]byte)}
}
