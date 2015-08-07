package envy

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type TestSimpleEnv map[string]string

func (t TestSimpleEnv) Read() map[string]string {
	return map[string]string(t)
}

func (t TestSimpleEnv) GetPrefix() string {
	return ""
}

// build a null routed logger to inject into go envy
var nullLogger = log.New(ioutil.Discard, "[goenvy] ", log.LstdFlags|log.Lshortfile)

func TestConfigFromSimpleEnv(t *testing.T) {
	expectedHost := "thepark"
	expectedPort := 4242
	expectedDebug := true
	expectedQuit := false
	expectedValues := []string{"foo", "bar"}
	expectedIntList := []int{100, 1, -2, 0, int(math.MaxInt64) - 1}

	var intListAsStringList []string
	for _, v := range expectedIntList {
		intListAsStringList = append(intListAsStringList, strconv.Itoa(v))
	}

	env := TestSimpleEnv{
		"HOST":      expectedHost,
		"PORT":      fmt.Sprintf("%d", expectedPort),
		"QUIT":      fmt.Sprintf("%v", expectedQuit),
		"DEBUG":     fmt.Sprintf("%v", expectedDebug),
		"SMTP_HOST": expectedHost,
		"VALUES":    strings.Join(expectedValues, ","),
		"INTLIST":   strings.Join(intListAsStringList, ","),
	}

	config := struct {
		Host     string
		Port     int
		Debug    bool
		Quit     bool
		SMTPHost string `name:"SMTP_HOST"`
		Values   []string
		IntList  []int
	}{}

	// function actually being tested
	err := LoadFromEnv(env, &config, nullLogger)
	if err != nil {
		t.Errorf("Error loading config: %s", err)
	}

	if config.Host != expectedHost {
		t.Errorf("Host was incorrect: got %q, expected %s", config.Host, expectedHost)
	}

	if config.Port != expectedPort {
		t.Errorf("Port was incorrect: got %d, expected %d", config.Port, expectedPort)
	}

	if config.Debug != expectedDebug {
		t.Errorf("Debug was incorrect: got %v, expected %v", config.Debug, expectedDebug)
	}

	if config.Quit != expectedQuit {
		t.Errorf("Quit was incorrect: got %v, expected %v", config.Quit, expectedQuit)
	}

	if config.SMTPHost != expectedHost {
		t.Errorf("SMTPHost was incorrect: got %q, expected %s", config.SMTPHost, expectedHost)
	}

	if !reflect.DeepEqual(config.Values, expectedValues) {
		t.Errorf("Values was incorrect: got %#v, expected %#v", config.Values, expectedValues)
	}

	if !reflect.DeepEqual(config.IntList, expectedIntList) {
		t.Errorf("IntList was incorrect: got %#v, expected %#v", config.IntList, expectedIntList)
	}

	t.Logf("config: %+v", config)
}

func TestShouldAllowTooManyConfigs(t *testing.T) {
	env := TestSimpleEnv{
		"HOST": "localhost",
		"PORT": "1234",
		"QUIT": "true",
	}

	config := struct {
		Host string
		Port int
	}{}

	// function actually being tested
	err := LoadFromEnv(env, &config, nullLogger)
	if err != nil {
		t.Errorf("Config should allow extra configs passed in")
	}
}

func TestShouldNotAllowTooFewConfigs(t *testing.T) {
	env := TestSimpleEnv{
		"HOST": "localhost",
	}

	config := struct {
		Host string
		Port int
	}{}

	// function actually being tested
	err := LoadFromEnv(env, &config, nullLogger)
	if err == nil {
		t.Errorf("Config should not allow too few configs passed in")
	}
}

func TestPassNonStructIn(t *testing.T) {
	env := TestSimpleEnv{}

	notStruct := "Im not a struct!"

	// function actually being tested
	err := LoadFromEnv(env, notStruct, nullLogger)
	if err == nil {
		t.Errorf("We should not process non struct references")
	}

	// function actually being tested
	err = LoadFromEnv(env, &notStruct, nullLogger)
	if err == nil {
		t.Errorf("We should not process a non struct")
	}

}

func TestAccessNonDefinedConfigKey(t *testing.T) {
	expectedHost := "thepark"

	env := TestSimpleEnv{
		"HOST": expectedHost,
	}

	config := struct {
		Host string
		Port int
	}{}

	// function actually being tested
	err := LoadFromEnv(env, &config, nullLogger)
	if err == nil {
		t.Errorf("Did not error when all values provided", err)
	}
}

func TestIntParsing(t *testing.T) {
	env := TestSimpleEnv{
		"PORT": "I'm a string!",
	}

	config := struct {
		Port int
	}{}

	// function actually being tested
	err := LoadFromEnv(env, &config, nullLogger)
	if err == nil {
		t.Errorf("Should have errored parsing string as int", err)
	}
}

func TestBoolParsing(t *testing.T) {
	env := TestSimpleEnv{
		"DEBUG": "Frak, I'm a string!",
	}

	config := struct {
		Debug bool
	}{}

	// function actually being tested
	err := LoadFromEnv(env, &config, nullLogger)
	if err == nil {
		t.Errorf("Should have errored parsing string as bool", err)
	}
}

// You should be able to parse an int list (comma-separated) to a string if the config structure says so
func TestIntListParsingAsString(t *testing.T) {
	expected := "100,2,-1,0,12"
	env := TestSimpleEnv{
		"INTLIST": expected,
	}
	config := struct {
		IntList string
	}{}

	err := LoadFromEnv(env, &config, nullLogger)
	if err != nil {
		t.Errorf("parsing string should not have errored out", err)
	}
	if expected != config.IntList {
		t.Errorf("Should have parsed intlist as string as specified")
	}
}

func TestStringListParsing(t *testing.T) {
	expected := "this,1,a,string,list"
	env := TestSimpleEnv{
		"STRLIST": expected,
	}
	config := struct {
		StrList []int
	}{}

	err := LoadFromEnv(env, &config, nullLogger)
	if err == nil {
		t.Errorf("Should have errored parsing string list as int list", err)
	}
}

// You shouldn't be able to use arrays in config struct
func TestStringArrayParsing(t *testing.T) {
	expected := "this,is,a,string"
	env := TestSimpleEnv{
		"STRLIST": expected,
	}
	config := struct {
		StrList [4]string
	}{}

	err := LoadFromEnv(env, &config, nullLogger)
	if err == nil {
		t.Errorf("Should have errored parsing into string array", err)
	}
}
