package simplecli

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// THE program name
var programName = os.Args[0]
var programArgs = os.Args[1:]

// OptionPrefix constant
const OptionPrefix = `--`

// Handle takes a point to a struct and construct a CommandGroup,
// based on the declared fields & methods of the specified struct.
func Handle(ptr interface{}) {
	handler := CommandGroup{MainCommand: ptr}
	handler.init(programArgs)
	handler.handle()
}

// CommandGroup group commands
type CommandGroup struct {
	MainCommand      interface{}
	CommandArgs      []string
	SubcommandByName map[string]reflect.Value
	SubCommandGroups []*CommandGroup
}

func (cmd *CommandGroup) init(commandArgs []string) {
	// Fixme: raise error if not pointer to struct
	typ := reflect.TypeOf(cmd.MainCommand).Elem()
	val := reflect.ValueOf(cmd.MainCommand).Elem()

	// Initialize subcommands with declared methods of cmd.MainCommand struct
	cmd.SubcommandByName = make(map[string]reflect.Value, typ.NumMethod())
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		cmd.SubcommandByName[method.Name] = val.MethodByName(method.Name)
	}

	// Initialize arguments and options
	cmd.CommandArgs = make([]string, 0)
	for i := range commandArgs {
		arg := commandArgs[i]

		if strings.HasPrefix(arg, OptionPrefix) {
			// If the argument starts with `--`, parse it as option.
			option := arg[len(OptionPrefix):]
			cmd.parseOption(option)
		} else {
			// TODO: add support for SubCommandGroups
			cmd.CommandArgs = append(cmd.CommandArgs, arg)
		}
	}
}

func (cmd *CommandGroup) parseOption(option string) {
	var field reflect.Value
	var val = reflect.ValueOf(cmd.MainCommand).Elem()

	// Parse option=value from string
	option, value := func() (string, string) {
		items := strings.Split(option, "=")
		if len(items) == 1 {
			return items[0], ""
		}

		return items[0], items[1]
	}()

	// Make sure there is a struct field with same name as `option``
	if field = val.FieldByName(strings.Title(option)); !field.IsValid() {
		message := fmt.Sprintf("The option --%s is not a recongized option", option)
		cmd.helpAndExit(-1, message)
	}

	// If there is a value passed, parse and set struct field
	if value != "" {
		parsedValue := cmd.parseAs(value, field.Kind())
		field.Set(parsedValue)
		return
	}

	// If no value passed, but for bool fields store true
	if field.Kind() == reflect.Bool {
		field.SetBool(true)
		return
	}

	// Raise error for non-Bool types if no value passed
	message := fmt.Sprintf("No value passed for option --%s", option)
	cmd.helpAndExit(-1, message)
}

func (cmd *CommandGroup) handle() {
	// No arguments passed
	if len(cmd.CommandArgs) <= 0 {
		cmd.helpAndExit(0)
	}

	// parse Args
	firstArg := cmd.CommandArgs[0]
	remainingArgs := cmd.CommandArgs[1:]

	// The first arg has corresponding receiver method on struct.
	method, ok := cmd.SubcommandByName[strings.Title(firstArg)]
	if !ok {
		message := fmt.Sprintf(" %s is not a valid command.", firstArg)
		cmd.helpAndExit(-1, message)
	}

	// The remaining arg types / len align with receiver method params on struct.
	methodType := method.Type()
	methodArgs := make([]reflect.Value, methodType.NumIn())
	if len(remainingArgs) != len(methodArgs) {
		message := fmt.Sprintf("%s requires %d argument(s).",
			firstArg,
			methodType.NumIn())
		cmd.helpAndExit(-1, message)
	}

	for i := 0; i < methodType.NumIn(); i++ {
		argI := methodType.In(i)
		methodArgs[i] = cmd.parseAs(remainingArgs[i], argI.Kind())
	}

	method.Call(methodArgs)
}

func (cmd *CommandGroup) parseAs(val string, kind reflect.Kind) reflect.Value {
	switch kind {
	case reflect.String:
		return reflect.ValueOf(val)

	case reflect.Int:
		num, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			message := fmt.Sprintf("%s is not a valid number.", val)
			cmd.helpAndExit(-1, message)
		}
		return reflect.ValueOf(int(num))

	case reflect.Bool:
		bool, err := strconv.ParseBool(val)
		if err != nil {
			message := fmt.Sprintf("%s is not a valid bool.", val)
			cmd.helpAndExit(-1, message)
		}
		return reflect.ValueOf(bool)
	default:
		message := fmt.Sprintf("Argument type %s is not supported yet.", kind)
		cmd.helpAndExit(-1, message)
		return reflect.ValueOf(nil)
	}
}

func (cmd *CommandGroup) helpAndExit(exitCode int, messages ...interface{}) {
	typ := reflect.TypeOf(cmd.MainCommand).Elem()
	val := reflect.ValueOf(cmd.MainCommand).Elem()

	for index := range messages {
		_, err := fmt.Fprintf(os.Stderr, "Error: %s\n", messages[index])
		if err != nil {
			panic(err)
		}
	}

	var prognInfo string
	hasOptions := val.NumField() > 0
	if hasOptions {
		prognInfo = fmt.Sprintf("Usage: %s [options] ", programName)
	} else {
		prognInfo = fmt.Sprintf("Usage: %s ", programName)
	}

	whitespaces := strings.Repeat(" ", len(prognInfo))
	cmdDescriptions := make([]string, 0, len(cmd.SubcommandByName))
	for k, m := range cmd.SubcommandByName {
		args := m.Type().String()
		desc := fmt.Sprintf("%s %s", strings.ToLower(k), args[4:])
		cmdDescriptions = append(cmdDescriptions, desc)
	}

	cmdHelp := strings.Join(cmdDescriptions, "\n"+whitespaces)
	help := prognInfo + cmdHelp

	if hasOptions {
		whitespaces = strings.Repeat(" ", 4)
		optDescriptions := make([]string, 0, typ.NumField())

		for i := 0; i < typ.NumField(); i++ {
			desc := fmt.Sprintf(
				"%s--%-10s %6s   %s",
				whitespaces,
				strings.ToLower(typ.Field(i).Name),
				typ.Field(i).Type,
				typ.Field(i).Tag,
			)
			optDescriptions = append(optDescriptions, desc)
		}

		optHelp := strings.Join(optDescriptions, "\n")
		help = help + "\nOptions:\n" + optHelp
	}

	fmt.Println(help)
	os.Exit(exitCode)
}
