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
func Handle(cmd interface{}) {
	grp := NewCommandGroup(cmd, programName, programArgs)
	defer grp.gracefulExit()

	grp.handle()
}

// CommandGroup group commands
type CommandGroup struct {
	MainCommand      interface{}
	CommandName      string
	CommandArgs      []string
	SubcommandByName map[string]reflect.Value
	SubCommandGroups []*CommandGroup
}

// NewCommandGroup constructs a new command group from a struct
func NewCommandGroup(cmd interface{}, commandName string, commandArgs []string) *CommandGroup {
	// Fixme: raise error if not pointer to struct
	typ := reflect.TypeOf(cmd).Elem()
	val := reflect.ValueOf(cmd).Elem()

	grp := CommandGroup{MainCommand: cmd}
	defer grp.gracefulExit()

	// Initialize subcommands with declared methods of cmd.MainCommand struct
	grp.SubcommandByName = make(map[string]reflect.Value, typ.NumMethod())
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		grp.SubcommandByName[method.Name] = val.MethodByName(method.Name)
	}

	// Initialize arguments and options
	grp.CommandName = commandName
	grp.CommandArgs = make([]string, 0)

	for i := range commandArgs {
		arg := commandArgs[i]

		if strings.HasPrefix(arg, OptionPrefix) {
			// If the argument starts with `--`, parse it as option.
			option := arg[len(OptionPrefix):]
			grp.parseOption(option)
		} else {
			// TODO: add support for SubCommandGroups
			grp.CommandArgs = append(grp.CommandArgs, arg)
		}
	}

	return &grp
}

func (grp *CommandGroup) gracefulExit() {
	if r := recover(); r != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", r)
		fmt.Fprintln(os.Stdout, grp.getHelp())
		os.Exit(-1)
	}
}

func (grp *CommandGroup) parseOption(option string) {
	var field reflect.Value
	var val = reflect.ValueOf(grp.MainCommand).Elem()

	// Parse option=value from string
	option, value := func() (string, string) {
		items := strings.Split(option, "=")
		if len(items) == 1 {
			return items[0], ""
		}

		return items[0], items[1]
	}()

	// Make sure there is a struct field with same name as `option`
	if field = val.FieldByName(strings.Title(option)); !field.IsValid() {
		message := fmt.Sprintf("The option --%s is not a recongized option", option)
		panic(message)
	}

	// If there is a value passed, parse and set struct field
	if value != "" {
		parsedValue := grp.parseAs(value, field.Kind())
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
	panic(message)
}

func (grp *CommandGroup) handle() {
	defer grp.gracefulExit()

	// if no arguments passed
	if len(grp.CommandArgs) <= 0 {
		message := fmt.Sprintf("No arguments passed for %s", grp.CommandName)
		panic(message)
	}

	subCommand := grp.CommandArgs[0]
	commandArgs := grp.CommandArgs[1:]

	// Check whether there is corresponding receiver method for subCommand
	method, ok := grp.SubcommandByName[strings.Title(subCommand)]
	if !ok {
		message := fmt.Sprintf(" %s is not a valid command.", subCommand)
		panic(message)
	}

	// The remaining arg types / len align with receiver method params on struct.
	methodType := method.Type()
	methodArgs := make([]reflect.Value, methodType.NumIn())

	if len(commandArgs) != len(methodArgs) {
		message := fmt.Sprintf("%s requires %d argument(s).",
			subCommand,
			methodType.NumIn())
		panic(message)
	}

	for i := 0; i < methodType.NumIn(); i++ {
		argI := methodType.In(i)
		methodArgs[i] = grp.parseAs(commandArgs[i], argI.Kind())
	}

	method.Call(methodArgs)
}

func (grp *CommandGroup) parseAs(val string, kind reflect.Kind) reflect.Value {
	switch kind {

	case reflect.String:
		return reflect.ValueOf(val)

	case reflect.Int:
		num, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			message := fmt.Sprintf("%s is not a valid number.", val)
			panic(message)
		}
		return reflect.ValueOf(int(num))

	case reflect.Bool:
		bool, err := strconv.ParseBool(val)
		if err != nil {
			message := fmt.Sprintf("%s is not a valid bool.", val)
			panic(message)
		}
		return reflect.ValueOf(bool)

	default:
		message := fmt.Sprintf("Argument type %s is currently not supported yet.", kind)
		panic(message)
	}
}

func (grp *CommandGroup) getHelp() string {
	typ := reflect.TypeOf(grp.MainCommand).Elem()
	val := reflect.ValueOf(grp.MainCommand).Elem()

	var prognInfo string
	hasOptions := val.NumField() > 0
	if hasOptions {
		prognInfo = fmt.Sprintf("Usage: %s [options] ", programName)
	} else {
		prognInfo = fmt.Sprintf("Usage: %s ", programName)
	}

	whitespaces := strings.Repeat(" ", len(prognInfo))
	cmdDescriptions := make([]string, 0, len(grp.SubcommandByName))
	for k, m := range grp.SubcommandByName {
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

	return help
}
