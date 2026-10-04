package model

import "testing"

func TestMySQLBit_ScanItoryVarchar(t *testing.T) {
	cases := []struct {
		in   any
		want MySQLBit
	}{{"1", 1}, {"0", 0}, {"", 0}, {nil, 0}, {[]uint8{1}, 1}, {int64(1), 1}, {true, 1}}
	for _, c := range cases {
		var b MySQLBit = 7
		if err := b.Scan(c.in); err != nil || b != c.want {
			t.Errorf("Scan(%#v) = %d, %v; want %d", c.in, b, err, c.want)
		}
	}
}
