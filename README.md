# SimpleCLI
The simplest way to handle cli arguments in golang. Inspired by python [Fire](https://github.com/google/python-fire) package.

* [Basic usage](#basic-usage)
* [With options / flags](#with-options--flags)
* [Nested commands](#nested-commands)

## Basic usage

```golang
package main

import (
        "fmt"
        "github.com/trhura/simplecli"
)

type Calc struct{}

// Need to be a public method, so that it is accessible by external package.
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

### CLI usage

```sh
thurahlaing @ simplecli > go build examples/simple/calc.go && ./calc
Usage: ./calc add (int, int)
              multiply (int, int)
```

```sh
thurahlaing @ simplecli > go build examples/simple/calc.go && ./calc divide
Error:  divide is not a valid command.
...
```

```sh
thurahlaing @ simplecli > go build examples/simple/calc.go && ./calc add 3
Error: add requires 2 argument(s).
...
```

```sh
thurahlaing @ simplecli > go build examples/simple/calc.go && ./calc add 3 as
Error: as is not a valid number.
...
```

```sh
thurahlaing @ simplecli > go build examples/simple/calc.go && ./calc add 3 4
7
```


## With options / flags

```golang
package main

import (
	"fmt"
	"github.com/trhura/simplecli"
	"strconv"
)

type Calc struct {
	// Need to be a public field, so that it is accessible by external package.
	Base    int  `base (radix) of input numbers`
	Verbose bool `print verbose output`
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
```

### CLI Usage

```sh
thurahlaing @ simplecli > go build examples/calc.go && ./calc
Usage: ./calc [options] add (int, int)
                        multiply (int, int)
Options:
    --base          int   base (radix) of input numbers
    --verbose      bool   print verbose output
```

```sh
thurahlaing @ simplecli > go build examples/simple/calc.go && ./calc add 01 10
11
```

```sh
thurahlaing @ simplecli > go build examples/simple/calc.go && ./calc --verbose add 01 10
1 + 10 = 11
```

```sh
thurahlaing @ simplecli > go build examples/simple/calc.go && ./calc --verbose --base=2 add 01 10
1 + 2 = 3
```

## Nested commands

```go
package main

import (
	"fmt"
	"github.com/trhura/simplecli"
)

// Database ...
type Database struct {
	Path string `database url path`
}

// Create database
func (db Database) Create() {
	fmt.Println("Creating database.")
}

// Drop database
func (db Database) Drop() {
	fmt.Println("Dropping database.")
}

// App ...
type App struct {
	Database *Database	// needs to be a pointer
	Port     int `server port `
}

// Start the app
func (app App) Start() {
	fmt.Printf("Listening app at %d.\n", app.Port)
}

// Reload the app
func (app App) Reload() {
	fmt.Println("Reloading app.")
}

// Kill the app
func (app App) Kill() {
	fmt.Println("Stoping app.")
}

func main() {
	simplecli.Handle(&App{
		Database: &Database{},
		Port:     8080,
	})
}
```

### CLI Usage

```sh
thurahlaing @ simplecli > go build examples/nested/app.go && ./app
Usage: ./app [options] kill ()
                       reload ()
                       start ()
                       database ...
Options:
        --port (int)		`server port`
```

```sh
thurahlaing @ simplecli >  go build examples/nested/app.go && ./app database
Usage: database [options] drop ()
                          create ()
Options:
        --path (string)		`database url path`
```

```sh
thurahlaing @ simplecli > go build examples/nested/app.go && ./app --port=80 start
Listening app at 80.
```

```sh
thurahlaing @ simplecli > go build examples/nested/app.go && ./app --port=80 database --path=dburl create
Creating database.
```
