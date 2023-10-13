package env

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

const (
	regexpString = "^[a-zA-Z_][a-zA-Z0-9_]*$"
)

type Value interface {
	String() string
	Set(string) error
}

type Getter interface {
	Value
	Get() any
}

// stringValue
type stringValue string

func newStringValue(p *string) *stringValue {
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() any { return string(*s) }

func (s *stringValue) String() string { return string(*s) }

// intValue
type intValue int64

func newIntValue(p *int64) *intValue {
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		return fmt.Errorf("error setting int value: %s", err)
	}
	*i = intValue(v)
	return nil
}

func (i *intValue) Get() any { return int64(*i) }

func (i *intValue) String() string { return strconv.FormatInt(int64(*i), 10) }

// env struct to handle env variables
type Env struct {
	Name  string
	Value Value
	isSet bool
}

func (e *Env) IsSet() bool {
	return e.isSet
}

type ErrorHandling int

const (
	ContinueOnError ErrorHandling = iota // Return a descriptive error.
	ExitOnError                          // Call os.Exit(2) or for -h/-help Exit(0).
	PanicOnError                         // Call panic with a descriptive error.
)

type EnvSet struct {
	name          string
	errorHandling ErrorHandling
	envs          map[string]*Env
	prefix        string
	parsed        bool
}

func isValidEnvironment(name string) (bool, error) {
	matched, err := regexp.MatchString(regexpString, name)
	return matched, err
}

func (e *EnvSet) Var(value Value, name string) error {
	var err error
	var env *Env
	var valid, alreadythere bool
	valid, err = isValidEnvironment(name)
	if err != nil {
		err = fmt.Errorf("Var: error while parsing env: %s with regexp: %s: %s", name, regexpString, err)
		goto handleErrors
	}
	if !valid {
		err = fmt.Errorf("Var: env name %s does not conform to regular expression: %s", name, regexpString)
		goto handleErrors
	}
	env = &Env{
		Name:  name,
		Value: value,
	}
	_, alreadythere = e.envs[name]
	if alreadythere {
		err = fmt.Errorf("%s flag redefined: %s", e.name, name)
	}

handleErrors:
	if err != nil {
		switch e.errorHandling {
		case ContinueOnError:
			return err
		case ExitOnError:
			fmt.Println(err.Error())
			os.Exit(2)
		case PanicOnError:
			panic(err)
		}
	}

	if e.envs == nil {
		e.envs = make(map[string]*Env)
	}
	e.envs[name] = env

	return nil
}

func NewEnvSet(name string, errorHandling ErrorHandling) *EnvSet {
	e := &EnvSet{
		name:          name,
		errorHandling: errorHandling,
	}
	return e
}

func (e *EnvSet) Init(name string, errorHandling ErrorHandling) {
	e.name = name
	e.errorHandling = errorHandling
}

var Environment = NewEnvSet(os.Args[0], ExitOnError)

func (e *EnvSet) SetPrefix(prefix string) error {
	valid, err := isValidEnvironment(prefix)
	if err != nil {
		err = fmt.Errorf("SetPrefix: error while parsing env prefix: %s with regexp: %s: %s", prefix, regexpString, err)
		goto handleErrors
	}
	if !valid {
		err = fmt.Errorf("SetPrefix: prefix %s does not conform to regular expression: %s", prefix, regexpString)
	}

handleErrors:
	if err != nil {
		switch e.errorHandling {
		case ContinueOnError:
			return err
		case ExitOnError:
			fmt.Println(err.Error())
			os.Exit(2)
		case PanicOnError:
			panic(err)
		}
	}

	e.prefix = prefix
	return nil
}

func SetPrefix(prefix string) error {
	return Environment.SetPrefix(prefix)
}

func (e *EnvSet) Int64Var(p *int64, name string) error {
	return e.Var(newIntValue(p), name)
}

func Int64Var(p *int64, name string) error {
	return Environment.Int64Var(p, name)
}

func (e *EnvSet) StringVar(p *string, name string) error {
	return e.Var(newStringValue(p), name)
}

func StringVar(p *string, name string) error {
	return Environment.StringVar(p, name)
}

func (e *EnvSet) Parse() error {
	for name, env := range e.envs {
		fullname := name
		if e.prefix != "" {
			fullname = e.prefix + "_" + name
		}
		var err error
		value := os.Getenv(fullname)
		if value != "" {
			err = env.Value.Set(value)
		}
		if err != nil {
			err = fmt.Errorf("Parse: error while parsing env: %s: %w", fullname, err)
			switch e.errorHandling {
			case ContinueOnError:
				return err
			case ExitOnError:
				fmt.Println(err.Error())
				os.Exit(2)
			case PanicOnError:
				panic(err)
			}
		} else {
			env.isSet = true
		}
	}
	e.parsed = true
	return nil
}

func Parse() error {
	return Environment.Parse()
}

func (e *EnvSet) Parsed() bool {
	return e.parsed
}

func Parsed() bool {
	return Environment.Parsed()
}

func (e *EnvSet) IsSet(name string) (bool, error) {
	env := e.envs[name]
	if env == nil {
		return false, fmt.Errorf("IsSet: env not defined: %s", name)
	}
	return env.IsSet(), nil
}

func IsSet(name string) (bool, error) {
	return Environment.IsSet(name)
}
