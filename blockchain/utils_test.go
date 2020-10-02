package blockchain

import (
	"fmt"
	"testing"
)

func TestNew32BytesArrWithRsh(t *testing.T) {
	cases := []struct {
		N            int
		FirstByteStr string
	}{
		{1, "10000000"},
		{2, "01000000"},
		{3, "00100000"},
	}
	for _, c := range cases {
		a := New32BytesArrWithRsh(c.N)
		s := fmt.Sprintf("%08b", a[0])
		if s != c.FirstByteStr {
			t.Fatalf("cal %q != exp %q with %d", s, c.FirstByteStr, c.N)
		} else {
			t.Logf("cal %q == exp %q with %d", s, c.FirstByteStr, c.N)
		}
	}
}

func TestCmp32BytesArr(t *testing.T) {
	cases := []struct {
		Ash      int
		Bsh      int
		Expected int
	}{
		{1, 2, 1},
		{2, 1, -1},
		{3, 3, 0},
	}
	for _, c := range cases {
		a := New32BytesArrWithRsh(c.Ash)
		b := New32BytesArrWithRsh(c.Bsh)
		if Cmp32BytesArr(a, b) != c.Expected {
			t.Fatalf("Cmp %d %d expected %d, got %d",
				c.Ash, c.Bsh, c.Expected, Cmp32BytesArr(a, b))
		}
	}
}
