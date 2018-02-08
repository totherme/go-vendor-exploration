package govendorexplorationb

import c "github.com/totherme/govendorexplorationc"

func BFunc() string {
	return c.CFunc()
}
