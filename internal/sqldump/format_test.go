package sqldump

import "testing"

func TestFormatSize(t *testing.T) {
	cases := []struct {
		size float64
		want string
	}{
		{0, "0.00 B"},
		{512, "512.00 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1024 * 1024, "1.00 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{1024 * 1024 * 1024 * 1024, "1.00 TB"},
		{1024 * 1024 * 1024 * 1024 * 1024, "1024.00 TB"}, // acima de TB não converte mais
	}
	for _, c := range cases {
		if got := FormatSize(c.size); got != c.want {
			t.Errorf("FormatSize(%v) = %q, want %q", c.size, got, c.want)
		}
	}
}
