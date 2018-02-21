# SimpleCLI
The simplest way to handle cli arguments in golang. Inspired by python [Fire](https://github.com/google/python-fire) package.

## Basic usage

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

### Example CLI usage

```sh
thurahlaing @ simplecli > go build calc.go && ./calc
Usage: ./calc add (int, int)
              multiply (int, int)
```

```sh
thurahlaing @ simplecli > go build main/calc.go && ./calc divide
Error:  Divide is not a valid command.
...
```

```sh
thurahlaing @ simplecli > go build main/calc.go && ./calc add 3
Error: add requires 2 argument(s).
...
```

```sh
thurahlaing @ simplecli > go build main/calc.go && ./calc add 3 as
Error: as is not a valid number.
...
```

```sh
thurahlaing @ simplecli > go build main/calc.go && ./calc add 3 4
7
```


## Advanced usage (with options / flags)

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

### Example CLI Usage

```sh
thurahlaing @ simplecli > go build main/calc.go && ./calc add 01 10
11
```

```sh
thurahlaing @ simplecli > go build main/calc.go && ./calc --verbose add 01 10
1 + 10 = 11
```

```sh
thurahlaing @ simplecli > go build main/calc.go && ./calc --verbose --base=2 add 01 10
1 + 2 = 3
```
