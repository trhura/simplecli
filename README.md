# SimpleCLI
The simplest way to handle cli arguments in golang. Inspired by python [Fire](https://github.com/google/python-fire) package. 

## Usage 

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
        simplecli.Handle(Calc{})
}
```

## Examples

```sh
> go build calc.go && ./calc
Usage: ./calc Multiply (int, int) |
              Add (int, int)
```

```sh
> go build calc.go && ./calc Add 3 5
8
```
