package main

import (
	"b"
	"c"
	"fmt"
)

func main() {
	thing := b.Thing{Content: c.OtherThing{Badgers: "Foo"}}
	fmt.Println("Running A: ", thing)
}
