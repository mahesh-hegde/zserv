package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert := assert.New(t)
	for in, out := range happyCases {
		got, err := parseLimit(in)
		assert.Nil(err)
		assert.Equal(out, got, "parseLimit("+in+")")
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
