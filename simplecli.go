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
		cli.helpAndExit(0)
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
		message := fmt.Sprintf(" %s is not a valid command.", firstArg)
		cli.helpAndExit(-1, message)
	}

	// The remaining arg types / len align with receiver method params on struct.
	methodType := method.Type()
	methodArgs := make([]reflect.Value, methodType.NumIn())
	if len(remainingArgs) != len(methodArgs) {
		message := fmt.Sprintf("%s requires %d argument(s).",
			firstArg,
			methodType.NumIn())
		cli.helpAndExit(-1, message)
	}

	for i := 0; i < methodType.NumIn(); i++ {
		argI := methodType.In(i)
		methodArgs[i] = cli.parseAs(remainingArgs[i], argI.Kind())
	}

	method.Call(methodArgs)
}

func (cli *cliHandler) parseAs(val string, kind reflect.Kind) reflect.Value {
	switch kind {
	case reflect.String:
		return reflect.ValueOf(val)

	case reflect.Int:
		num, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			message := fmt.Sprintf("%s is not a valid number.", val)
			cli.helpAndExit(-1, message)
		}
		return reflect.ValueOf(int(num))

	case reflect.Bool:
		bool, err := strconv.ParseBool(val)
		if err != nil {
			message := fmt.Sprintf("%s is not a valid bool.", val)
			cli.helpAndExit(-1, message)
		}
		return reflect.ValueOf(bool)
	default:
		message := fmt.Sprintf("Argument type %s is not supported yet.", kind)
		cli.helpAndExit(-1, message)
		return reflect.ValueOf(nil)
	}
}

func (cli *cliHandler) helpAndExit(exitCode int, messages ...interface{}) {
	for index := range messages {
		_, err := fmt.Fprintf(os.Stderr, "Error: %s\n", messages[index])
		if err != nil {
			panic(err)
		}
	}

	prognInfo := fmt.Sprintf("Usage: %s ", cli.prgn)
	whitespaces := strings.Repeat(" ", len(prognInfo))

	cmdDescriptions := make([]string, 0, len(cli.methodsByName))
	for k, m := range cli.methodsByName {
		desc := fmt.Sprintf("%s %s", k, m.Type().String()[4:])
		cmdDescriptions = append(cmdDescriptions, desc)
	}

	commandsHelp := strings.Join(cmdDescriptions, "\n"+whitespaces)
	help := prognInfo + commandsHelp

	fmt.Println(help)
	os.Exit(exitCode)
}
