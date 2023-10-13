package env_test

import (
	"strconv"
	"testing"

	"github.com/juweissenberg/simpleconf/pkg/env"
	"gotest.tools/v3/assert"
	envtools "gotest.tools/v3/env"
)

const (
	testInt64EnvName   string = "TEST_INT64"
	testInt64EnvValue  int64  = int64(42)
	testStringEnvName  string = "TEST_STRING"
	testStringEnvValue string = "forty-two"
)

func _TestEnv(t *testing.T) {

	var testInt64 int64
	{
		err := env.Int64Var(&testInt64, testInt64EnvName)
		assert.NilError(t, err)
	}

	var testString string
	{
		err := env.StringVar(&testString, testStringEnvName)
		assert.NilError(t, err)
	}

	err := env.Parse()
	assert.NilError(t, err)
	assert.Assert(t, env.Parsed())

	{
		isSet, err := env.IsSet(testInt64EnvName)
		assert.NilError(t, err)
		assert.Assert(t, isSet)
	}

	{
		isSet, err := env.IsSet(testStringEnvName)
		assert.NilError(t, err)
		assert.Assert(t, isSet)
	}

	assert.Equal(t, testInt64, testInt64EnvValue)
	assert.Equal(t, testString, testStringEnvValue)
}

func TestEnv(t *testing.T) {
	env.Environment = env.NewEnvSet("test", env.ContinueOnError)

	defer envtools.PatchAll(t, map[string]string{
		testInt64EnvName:  strconv.FormatInt(testInt64EnvValue, 10),
		testStringEnvName: testStringEnvValue,
	})()

	_TestEnv(t)
}

func TestEnvWithPrefix(t *testing.T) {
	env.Environment = env.NewEnvSet("testWithPrefix", env.ContinueOnError)

	const prefix = "PREFIX"
	env.SetPrefix(prefix)

	defer envtools.PatchAll(t, map[string]string{
		prefix + "_" + testInt64EnvName:  strconv.FormatInt(testInt64EnvValue, 10),
		prefix + "_" + testStringEnvName: testStringEnvValue,
	})()

	_TestEnv(t)
}

func TestEnvWithErrors(t *testing.T) {
	env.Environment = env.NewEnvSet("testWithErrors", env.ContinueOnError)

	const badFormat = "%.?/$"
	if err := env.SetPrefix(badFormat); err == nil {
		t.FailNow()
	}

	defer envtools.PatchAll(t, map[string]string{
		testInt64EnvName: "not_a_number",
	})()

	var testInt64 int64
	{
		err := env.Int64Var(&testInt64, testInt64EnvName)
		assert.NilError(t, err)
	}

	var testInt64IllegalName int64
	{
		err := env.Int64Var(&testInt64IllegalName, badFormat)
		if err == nil {
			t.FailNow()
		}
	}

	var testInt64DefinedTwice int64
	{
		err := env.Int64Var(&testInt64DefinedTwice, "DUPLICATE")
		assert.NilError(t, err)
		err = env.Int64Var(&testInt64DefinedTwice, "DUPLICATE")
		if err == nil {
			t.FailNow()
		}
	}

	{
		_, err := env.IsSet("UNDEFINED_ENV")
		if err == nil {
			t.FailNow()
		}
	}

	if err := env.Parse(); err == nil {
		t.FailNow()
	}
}

func TestEnvEdgeCases(t *testing.T) {
	env.Environment = env.NewEnvSet("testEdgeCases", env.ContinueOnError)
	env.Environment.Init("testEdgeCases", env.ContinueOnError)
}
