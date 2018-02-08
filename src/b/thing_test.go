package b_test

import (
	"b"
	"testing"
)

func TestBFunc(t *testing.T) {
	if b.BFunc() != "Badgers" {
		t.Fail()
	}
}
