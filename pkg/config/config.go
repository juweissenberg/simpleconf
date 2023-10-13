package config

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	"github.com/juweissenberg/simpleconf/pkg/env"
)

type config interface{}

type parser struct {
	flagSet *flag.FlagSet
	envSet  *env.EnvSet
	config  config
}

func NewParser(config config) *parser {
	name := reflect.TypeOf(config).Elem().Name()
	c := &parser{
		flagSet: flag.NewFlagSet(name, flag.ContinueOnError),
		envSet:  env.NewEnvSet(name, env.ContinueOnError),
		config:  config,
	}
	return c
}

func (p *parser) SetEnvPrefix(prefix string) *parser {
	p.envSet.SetPrefix(prefix)
	return p
}

func (p *parser) Parse() error {
	configType := reflect.TypeOf(p.config).Elem()
	configValue := reflect.ValueOf(p.config).Elem()
	count := configType.NumField()

	for i := 0; i < count; i++ {
		field := configType.Field(i)
		usage := field.Tag.Get("usage")

		flag := field.Tag.Get("arg")
		env := field.Tag.Get("env")

		fieldValue := configValue.FieldByName(field.Name)

		switch field.Type.Kind() {
		case reflect.String:
			defaultValue := fieldValue.String()
			pointer := fieldValue.Addr().Interface().(*string)
			if flag != "" {
				p.flagSet.StringVar(pointer, flag, defaultValue, usage)
			}
			if env != "" {
				p.envSet.StringVar(pointer, env)
			}
		case reflect.Int64:
			defaultValue := fieldValue.Int()
			pointer := fieldValue.Addr().Interface().(*int64)
			if flag != "" {
				p.flagSet.Int64Var(pointer, flag, defaultValue, usage)
			}
			if env != "" {
				p.envSet.Int64Var(pointer, env)
			}
		default:
			panic(fmt.Errorf("parsing environment of type %s is not implemented", field.Type.Kind()))
		}
	}

	err := p.flagSet.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	err = p.envSet.Parse()
	if err != nil {
		return err
	}

	return nil
}
