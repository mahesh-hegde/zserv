package main

import (
	"errors"
	"testing"
)

func TestParseLimit(t *testing.T) {
	happyCases := map[string]int64{
		"1500":  1500,
		"1M":    1024 * 1024,
		"250M":  250 * 1024 * 1024,
		"1K":    1024,
		"0K":    0,
		"1500K": 1500 * 1024,
		"1G":    1024 * 1024 * 1024,
		"23G":   1024 * 1024 * 1024 * 23,
	}
	for in, out := range happyCases {
		got, err := parseLimit(in)
		if err != nil {
			t.Errorf("parseLimit(%s): Expected %d but got error %s",
				in, out, err.Error())
		}
		if got != out {
			t.Errorf("parseLimit(%s): Expected %d but got %d", in, out, got)
		}
	}
}

func TestParseLimitErrors(t *testing.T) {
	invalidFormat := []string{"70T", "600KK", "6001L", "K", "M", "G"}
	for _, in := range invalidFormat {
		_, err := parseLimit(in)
		if !errors.Is(err, errLimitFormat) {
			t.Errorf("Expected errLimitFormat (%s), but got %s",
				errLimitFormat.Error(), err.Error())
		}
	}
}
