package netest

import (
	"fmt"
)

const (
	Kilobyte = 1024.0
	Megabyte = 1048576.0
	Gigabyte = 1073741824.0
)

func Humanize(v float64) string {
	if v >= Gigabyte {
		return fmt.Sprintf("%.1f GB", (v / Gigabyte))
	} else if v >= Megabyte {
		return fmt.Sprintf("%.1f MB", (v / Megabyte))
	} else if v >= Kilobyte {
		return fmt.Sprintf("%.1f KB", (v / Kilobyte))
	} else {
		return fmt.Sprintf("%.1f B", v)
	}
}
