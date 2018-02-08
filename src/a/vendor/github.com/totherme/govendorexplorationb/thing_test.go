package govendorexplorationb_test

import (
	"testing"

	b "github.com/totherme/govendorexplorationb"
)

func TestBFunc(t *testing.T) {
	if b.BFunc() != "Ostriches" {
		t.Fail()
	}
}
