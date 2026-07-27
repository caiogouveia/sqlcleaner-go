package sqldump

import "fmt"

// FormatSize formata um tamanho em bytes como B/KB/MB/GB/TB, igual ao format_size() do Python.
func FormatSize(size float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	for i, unit := range units {
		if size < 1024.0 || i == len(units)-1 {
			return fmt.Sprintf("%.2f %s", size, unit)
		}
		size /= 1024.0
	}
	return fmt.Sprintf("%.2f %s", size, units[len(units)-1])
}
