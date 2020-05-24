package main

import (
	"fmt"
	"strconv"
	"testing"
)

func Test_printLargeNumber(t *testing.T) {
	if printLargeNumber(0) != "0" {
		t.Error(fmt.Sprintf("0 != %s", printLargeNumber(0)))
	}
	if printLargeNumber(1234) != "1234" {
		t.Error(fmt.Sprintf("1234 != %s", printLargeNumber(1234)))
	}
	if printLargeNumber(1234) != "1234" {
		t.Error(fmt.Sprintf("1234 != %s", printLargeNumber(1234)))
	}
	if printLargeNumber(10000) != "10 000" {
		t.Error(fmt.Sprintf("10 000 != %s", printLargeNumber(10000)))
	}
	if printLargeNumber(1234000) != "1 234 000" {
		t.Error(fmt.Sprintf("1 234 000 != %s", printLargeNumber(1234000)))
	}
}

func betsAllowedTestHelper(t *testing.T, hour int, expected bool) {
	if betsAllowed(hour, 0) != expected {
		t.Error(fmt.Sprintf("betsAllowed(%d, %d) != %s", hour, 0, strconv.FormatBool(expected)))
	}
}

func Test_betsAllowed(t *testing.T) {
	betsAllowedTestHelper(t, 0, true)
	betsAllowedTestHelper(t, betTimeTo-1, true)
	betsAllowedTestHelper(t, betTimeTo, true)
	betsAllowedTestHelper(t, betTimeTo+1, false)
	betsAllowedTestHelper(t, betTimeFrom-1, false)
	betsAllowedTestHelper(t, betTimeFrom, true)
	betsAllowedTestHelper(t, betTimeFrom+1, true)
	betsAllowedTestHelper(t, 23, true)
}
