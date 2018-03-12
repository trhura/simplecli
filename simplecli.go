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
func Handle(command interface{}) {
	cmdGroup := newCommandGroup(programName, command)
	defer cmdGroup.gracefulExit()

	cmdGroup.handle(programArgs)
}

// CommandGroup group commands
type CommandGroup struct {
	command          interface{}
	commandName      string
	subCommandByName map[string]reflect.Value
	subCommandGroups map[string]*CommandGroup
}

// newCommandGroup constructs a new command group from a struct pointer
func newCommandGroup(name string, ptr interface{}) *CommandGroup {
	ptrVal := reflect.ValueOf(ptr)
	if !isPtrToStruct(ptrVal) {
		panic("The passed interface is not a pointer to a struct.")
	}

	structVal := reflect.ValueOf(ptr).Elem()
	structTyp := reflect.TypeOf(ptr).Elem()

	cmdGroup := CommandGroup{
		command:     ptr,
		commandName: name,
	}

	// Initialize subcommands with declared methods of cmd.Command struct
	cmdGroup.subCommandByName = make(map[string]reflect.Value)
	for i := 0; i < structVal.NumMethod(); i++ {
		methodName := structTyp.Method(i).Name
		lowerName := strings.ToLower(methodName)
		cmdGroup.subCommandByName[lowerName] = structVal.MethodByName(methodName)
	}

	// Scan structs fields for possible SubCommands
	cmdGroup.subCommandGroups = make(map[string]*CommandGroup)
	for i := 0; i < structVal.NumField(); i++ {
		fieldTyp := structTyp.Field(i)
		fieldVal := structVal.Field(i)
		if isPtrToStruct(fieldVal) {
			lowerName := strings.ToLower(fieldTyp.Name)
			cmdGroup.subCommandGroups[lowerName] = newCommandGroup(
				lowerName, fieldVal.Interface(),
			)
		}
	}

	return &cmdGroup
}

func (grp *CommandGroup) gracefulExit() {
	if message := recover(); message != nil {
		fmt.Fprint(os.Stderr, message)
		fmt.Fprintln(os.Stdout, grp.getHelp())
		os.Exit(-1)
	}
}

func (grp *CommandGroup) parseOption(option string) {
	var field reflect.Value
	var reflectedValue = reflect.ValueOf(grp.command).Elem()

	// Parse option=value from string
	option, value := func() (string, string) {
		items := strings.Split(option, "=")
		if len(items) == 1 {
			return items[0], ""
		}

		return items[0], items[1]
	}()

	// Make sure there is a struct field with same name as `option`
	if field = reflectedValue.FieldByName(strings.Title(option)); !field.IsValid() {
		message := fmt.Sprintf("The option --%s is not a recongized option.\n", option)
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
	message := fmt.Sprintf("No value passed for option --%s.\n", option)
	panic(message)
}

// invoke either relevant method in CommandGroup or chain to SubCommandGroup
func (grp *CommandGroup) handle(optargs []string) {
	defer grp.gracefulExit()

	var args []string
	for i, arg := range optargs {
		// parse options as long as it starts with `--`
		if strings.HasPrefix(arg, OptionPrefix) {
			option := arg[len(OptionPrefix):]
			grp.parseOption(option)
		} else {
			args = optargs[i:]
			break
		}
	}

	// if no arguments passed
	if len(args) <= 0 {
		panic("")
	}

	subcmd := args[0]
	subargs := args[1:]

	if subgrp, ok := grp.subCommandGroups[subcmd]; ok {
		subgrp.handle(subargs)
		return
	}

	// Check whether there is corresponding receiver method for subCommand
	method, ok := grp.subCommandByName[subcmd]
	if !ok {
		message := fmt.Sprintf(" %s is not a valid command for %s.\n", subcmd, grp.commandName)
		panic(message)
	}

	// The remaining arg types / len align with receiver method params on struct.
	methodType := method.Type()
	methodArgs := make([]reflect.Value, methodType.NumIn())

	if len(subargs) != len(methodArgs) {
		message := fmt.Sprintf("%s requires %d argument(s).\n",
			subcmd,
			methodType.NumIn())
		panic(message)
	}

	for i := 0; i < methodType.NumIn(); i++ {
		argIn := methodType.In(i)
		methodArgs[i] = grp.parseAs(subargs[i], argIn.Kind())
	}

	method.Call(methodArgs)
}

// parse cli arg (string) to specific golang type
func (grp *CommandGroup) parseAs(arg string, kind reflect.Kind) reflect.Value {
	switch kind {

	case reflect.String:
		return reflect.ValueOf(arg)

	case reflect.Int:
		num, err := strconv.ParseInt(arg, 10, 32)
		if err != nil {
			message := fmt.Sprintf("%s is not a valid number.\n", arg)
			panic(message)
		}
		return reflect.ValueOf(int(num))

	case reflect.Bool:
		bool, err := strconv.ParseBool(arg)
		if err != nil {
			message := fmt.Sprintf("%s is not a valid bool.\n", arg)
			panic(message)
		}
		return reflect.ValueOf(bool)

	default:
		message := fmt.Sprintf("Argument type %s is currently not suppotyped yet.\n", kind)
		panic(message)
	}
}

func (grp *CommandGroup) getHelp() string {
	typ := reflect.TypeOf(grp.command).Elem()
	val := reflect.ValueOf(grp.command).Elem()

	numOptions := val.NumField() - len(grp.subCommandGroups)
	hasOptions := numOptions > 0

	var programInfo string
	if hasOptions {
		programInfo = fmt.Sprintf("Usage: %s [options] ", grp.commandName)
	} else {
		programInfo = fmt.Sprintf("Usage: %s ", grp.commandName)
	}

	indentation := strings.Repeat(" ", len(programInfo))
	descriptions := make([]string, 0, len(grp.subCommandByName))

	for name, cmd := range grp.subCommandByName {
		args := cmd.Type().String()[4:] // 4 is the magic number to strip `Func`
		desc := fmt.Sprintf("%s %s", name, args)
		descriptions = append(descriptions, desc)
	}

	for name := range grp.subCommandGroups {
		desc := fmt.Sprintf("%s ...", name)
		descriptions = append(descriptions, desc)
	}

	help := programInfo + strings.Join(descriptions, "\n"+indentation)

	if hasOptions {
		indentation = strings.Repeat(" ", 7)
		descriptions := make([]string, 0, numOptions)

		for i := 0; i < typ.NumField(); i++ {
			if !isPtrToStruct(val.Field(i)) {
				field := typ.Field(i)
				format := "%s --%s (%s)		`%s`"
				desc := fmt.Sprintf(format, indentation, strings.ToLower(field.Name), field.Type, field.Tag)
				descriptions = append(descriptions, desc)
			}
		}

		options := "Options:\n" + strings.Join(descriptions, "\n")
		help = help + "\n" + options
	}

	return help
}

// return whether the value is a pointer to a struct
func isPtrToStruct(v reflect.Value) bool {
	return v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct
}
