package main

import (
	"fmt"
	"github.com/trhura/simplecli"
	"strconv"
)

type Calc struct {
	Base    int // Need to be a public field, so that it is accessible by external package.
	Verbose bool
}

// Need to be a public method, so that it is accessible by external package.
func (c Calc) Add(x int, y int) {
	xb, _ := strconv.ParseInt(strconv.Itoa(x), c.Base, 32)
	yb, _ := strconv.ParseInt(strconv.Itoa(y), c.Base, 32)
	if c.Verbose {
		fmt.Printf("%d + %d = ", xb, yb)
	}
	fmt.Println(xb + yb)
}

func (c Calc) Multiply(x int, y int) {
	xb, _ := strconv.ParseInt(strconv.Itoa(x), c.Base, 32)
	yb, _ := strconv.ParseInt(strconv.Itoa(y), c.Base, 32)
	if c.Verbose {
		fmt.Printf("%d * %d = ", xb, yb)
	}
	fmt.Println(xb * yb)
}

func main() {
	// Needs to pass a pointer to struct, so it can be modified.
	simplecli.Handle(&Calc{Base: 10, Verbose: false})
}
