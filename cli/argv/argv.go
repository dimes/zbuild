// Package argv is a slight improvement over the Go flags package for parsing
// arguments of nested commands
package argv

import (
	"builder/buildlog"
	"errors"
	"fmt"
	"regexp"
)

type argType string

const (
	stringType argType = "string"
	boolType   argType = "bool"
)

var (
	argNameRegex = regexp.MustCompile("^-[0-9A-Za-z]+$")
)

// ArgSet is a set of arguments to be parsed
type ArgSet struct {
	parsed         bool
	registeredArgs map[string]*arg
}

type arg struct {
	name        string
	argType     argType
	ptr         interface{}
	parsed      bool
	description string
}

// NewArgSet returns an array of args
func NewArgSet() *ArgSet {
	return &ArgSet{
		registeredArgs: make(map[string]*arg),
	}
}

// Parse parses the passed in arguments. The returned array is the left-over arguments.
func (a *ArgSet) Parse(args []string) ([]string, error) {
	var argForValue *arg
	rest := make([]string, 0)
	for _, argString := range args {
		if argForValue == nil {
			// In this branch we are parsing the arg name
			if !argNameRegex.MatchString(argString) {
				rest = append(rest, argString)
				continue
			}

			argName := argString[1:]
			if arg, ok := a.registeredArgs[argName]; ok && !arg.parsed {
				arg.parsed = true
				if arg.argType == boolType {
					*(arg.ptr.(*bool)) = true
				} else {
					argForValue = arg
				}
			} else {
				rest = append(rest, argString)
			}
		} else {
			// In this branch we are parsing the arg value
			switch argForValue.argType {
			case stringType:
				*(argForValue.ptr.(*string)) = argString
			default:
				return nil, fmt.Errorf("Unknown arg type %s", argForValue.argType)
			}
			argForValue = nil
		}
	}

	a.parsed = true
	return rest, nil
}

// PrintUsage prints the usage of this arg set to std in
func (a *ArgSet) PrintUsage() {
	for argName, arg := range a.registeredArgs {
		buildlog.Infof("\t-%s\t%s\n", argName, arg.description)
	}
}

// ExpectString expects a string argument with the given name.
func (a *ArgSet) ExpectString(ptr *string, name, defaultValue, description string) error {
	*ptr = defaultValue
	return a.expectArg(name, stringType, ptr, description)
}

// ExpectBool expects a bool argument with the given name
func (a *ArgSet) ExpectBool(ptr *bool, name string, defaultValue bool, description string) error {
	*ptr = defaultValue
	return a.expectArg(name, boolType, ptr, description)
}

func (a *ArgSet) expectArg(name string, argType argType, ptr interface{}, description string) error {
	if a.parsed {
		return errors.New("Parse has already been called on this instance")
	}

	if _, ok := a.registeredArgs[name]; ok {
		return fmt.Errorf("Already expecting arg for %s", name)
	}

	a.registeredArgs[name] = &arg{
		name:        name,
		argType:     argType,
		ptr:         ptr,
		description: description,
	}

	return nil
}
