package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// parse memory limit strings like "256K" or "4M"

var limitRegex = regexp.MustCompile("([0-9]+)[KMG]")

var multipliers = map[rune]int64{
	'K': 1024,
	'M': 1024 * 1024,
	'G': 1024 * 1024 * 1024,
}

func parseLimit(limit string) (bytes int64, err error) {
	bytes, err = strconv.ParseInt(limit, 10, 64)
	if !errors.Is(err, strconv.ErrSyntax) {
		return bytes, err
	}
	if limitRegex.FindString(limit) == limit {
		var mult rune
		fmt.Sscanf(limit, "%d%c", &bytes, &mult)
		return bytes * multipliers[mult], nil
	}
	return 0, fmt.Errorf("limit string is in invalid format")
}
