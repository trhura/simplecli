package simplecli

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Handle takes a struct type and call the relevant method of
// the struct based on the cli arguments passed.
func Handle(any interface{}) {
	handler := cliHandler{any: any}
	handler.init()
	handler.handle()
}

type cliHandler struct {
	any           interface{}
	typ           reflect.Type
	val           reflect.Value
	prgn          string
	args          []string
	methodsByName map[string]reflect.Value
}

func (cli *cliHandler) init() {
	// Must be a struct
	if reflect.ValueOf(cli.any).Kind() != reflect.Struct {
		panic(fmt.Sprintf("The argument must be a struct, got a %T instead.", cli.any))
	}

	// Init Fields
	cli.typ = reflect.TypeOf(cli.any)
	cli.val = reflect.ValueOf(cli.any)
	cli.methodsByName = make(map[string]reflect.Value, cli.typ.NumMethod())

	// Init Methods
	for i := 0; i < cli.typ.NumMethod(); i++ {
		method := cli.typ.Method(i)
		cli.methodsByName[method.Name] = cli.val.MethodByName(method.Name)
	}

	// Init Args
	cli.prgn = os.Args[0]
	cli.args = os.Args[1:]
	if len(cli.args) <= 0 {
		cli.printHelp()
		os.Exit(0)
	}

	// TODO: Init Defaults in the struct
}

func (cli *cliHandler) handle() {
	// parse Args
	firstArg := cli.args[0]
	remainingArgs := cli.args[1:]

	// The first arg has corresponding receiver method on struct.
	method, ok := cli.methodsByName[firstArg]
	if !ok {
		errMsg := fmt.Sprintf("Illegal command -- %s.", firstArg)
		panic(errMsg)
	}

	// The remaining arg types / len align with receiver method params on struct.
	methodType := method.Type()
	methodArgs := make([]reflect.Value, methodType.NumIn())
	if len(remainingArgs) != len(methodArgs) {
		errMsg := fmt.Sprintf("Illegal options -- %s takes %d argument(s).", firstArg, methodType.NumIn())
		panic(errMsg)
	}

	for i := 0; i < methodType.NumIn(); i++ {
		argI := methodType.In(i)
		methodArgs[i] = parseAs(remainingArgs[i], argI.Kind())
	}

	method.Call(methodArgs)
}

func parseAs(val string, kind reflect.Kind) reflect.Value {
	switch kind {
	case reflect.String:
		return reflect.ValueOf(val)

	case reflect.Int:
		num, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			errMsg := fmt.Sprintf("Illegal arg --  %s is not a valid int.", val)
			panic(errMsg)
		}
		return reflect.ValueOf(int(num))

	case reflect.Bool:
		bool, err := strconv.ParseBool(val)
		if err != nil {
			errMsg := fmt.Sprintf("Illegal arg --  %s is not a valid bool.", val)
			panic(errMsg)
		}
		return reflect.ValueOf(bool)
	default:
		errMsg := fmt.Sprintf("Illegal arg -- argument type %s is not supported.", kind)
		panic(errMsg)
	}
}

func (cli *cliHandler) printHelp() {
	prognInfo := fmt.Sprintf("Usage: %s ", cli.prgn)
	whitespaces := strings.Repeat(" ", len(prognInfo))

	cmdDescriptions := make([]string, 0, len(cli.methodsByName))
	for k, m := range cli.methodsByName {
		desc := fmt.Sprintf("%s %s", k, m.Type().String()[4:])
		cmdDescriptions = append(cmdDescriptions, desc)
	}

	commandsHelp := strings.Join(cmdDescriptions, " |\n"+whitespaces)
	help := prognInfo + commandsHelp
	fmt.Println(help)
}
