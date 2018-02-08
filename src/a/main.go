package main

import (
	"fmt"

	b "github.com/totherme/govendorexplorationb"
	c "github.com/totherme/govendorexplorationc"
)

func main() {
	fmt.Println("Running A: ", b.BFunc(), c.CFunc())
}
