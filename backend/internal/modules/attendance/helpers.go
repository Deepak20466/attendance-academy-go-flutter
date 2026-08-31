package attendance

import (
	"errors"
	"strconv"
	"strings"
)

var ErrNoCheckIn = errors.New("no matching check-in found for check-out")

func itoaLocal(n int) string { return strconv.Itoa(n) }

func joinLocal(parts []string, sep string) string { return strings.Join(parts, sep) }

func atoiDefault(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}
