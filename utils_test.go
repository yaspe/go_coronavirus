package main

import (
	"fmt"
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
