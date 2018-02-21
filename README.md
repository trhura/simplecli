# SimpleCLI
The simplest way to handle cli arguments in golang. Inspired by python [Fire](https://github.com/google/python-fire) package. 

## Basic Sample 

```golang
package main

import (
        "fmt"
        "github.com/trhura/simplecli"
)

type Calc struct{}

func (c Calc) Add(x int, y int) {
        fmt.Println(x + y)
}

func (c Calc) Multiply(x int, y int) {
        fmt.Println(x * y)
}

func main() {
        simplecli.Handle(&Calc{})
}
```

### Example Usage

```sh
thurahlaing @ simplecli > go build calc.go && ./calc
Usage: ./calc Add (int, int)
              Multiply (int, int)
```

```sh
thurahlaing @ simplecli > go build calc.go && ./calc Add 3 5
8
```

## With flags 

```golang
package main

import (
	"fmt"
	"github.com/trhura/simplecli"
	"strconv"
)

type Calc struct {
	Base    int
	Verbose bool
}

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
	simplecli.Handle(&Calc{Base: 10, Verbose: false})
}
```

### Example Usage

```sh
thurahlaing @ simplecli > go build main/calc.go && ./calc Add 01 10
11
```

```sh
thurahlaing @ simplecli > go build main/calc.go && ./calc --verbose Add 01 10
1 + 10 = 11
```

```sh
thurahlaing @ simplecli > go build main/calc.go && ./calc --verbose --base=2 Add 01 10
1 + 2 = 3
```
