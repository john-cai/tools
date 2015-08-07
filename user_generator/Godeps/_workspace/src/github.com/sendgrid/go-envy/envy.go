package envy

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var (
	// Error used when an invalid reference is provided to the Load function
	ErrInvalidConfigType = errors.New("A struct reference is required for loading config")

	// Basic config error
	ErrConfigInvalid = errors.New("Config is invalid")
)

var defaultLogger = log.New(os.Stderr, "[goenvy] ", log.LstdFlags|log.Lshortfile)

var SeparatorCharacter = ","

// Loads the config using a prefix
func LoadWithPrefix(prefix string, spec interface{}) error {
	// Helper function for using a prefix: Not Testable
	osEnv := &OsEnvironmentReader{Prefix: prefix}
	return LoadFromEnv(osEnv, spec, nil)
}

// Loads config from the provided EnvironmentReader
func LoadFromEnv(reader EnvironmentReader, configSpec interface{}, logger *log.Logger) error {
	if logger == nil {
		logger = defaultLogger
	}
	source := reader.Read()

	// make sure that we have something with which we can work
	if reflect.ValueOf(configSpec).Kind() != reflect.Ptr {
		return ErrInvalidConfigType
	}

	// Find the value of the provided configSpec
	// It must be a struct of some kind in order for the values
	// to be set.
	s := reflect.ValueOf(configSpec).Elem()
	if s.Kind() != reflect.Struct {
		return ErrInvalidConfigType
	}

	// create a list of all errors
	errors := make([]error, 0)

	// iterate over all fields in the struct
	typeOfSpec := s.Type()

	// make sure that we got the right number of configs
	expectedConfigCount := s.NumField()
	actualConfigCount := len(source)

	if expectedConfigCount > actualConfigCount {
		err := fmt.Errorf("Unexpected number of config values. Got %d, expected %d",
			actualConfigCount, expectedConfigCount)
		errors = append(errors, err)
	}

	// reflect on config values and set them to the right types
	for i := 0; i < s.NumField(); i++ {
		// reference to the value of the field (used for assignment)
		fieldValue := s.Field(i)
		// reference to the type of the field
		// (used to determine the name and any relevant struct tags)
		fieldType := typeOfSpec.Field(i)

		// Only uppercase values can be set (limitation of reflection)
		if fieldValue.CanSet() {
			fieldName := fieldType.Name

			// always assume uppercase key names
			key := strings.ToUpper(fieldName)

			overrideName := fieldType.Tag.Get("name")
			if overrideName != "" {
				key = overrideName
			}

			// string used for outputting useful error messages
			example := fieldType.Tag.Get("example")

			prefixedKey := fmt.Sprintf("%s%s", reader.GetPrefix(), key)

			// retrieve the value from the source, UPCASED
			// if this value is not available, track the error and continue with
			// the other options
			value, ok := source[key]
			if !ok {
				err := fmt.Errorf("Config not found: key=%s; example=%q", prefixedKey, example)
				errors = append(errors, err)
				continue
			}

			// populate the struct values based on what type it is
			switch fieldValue.Kind() {
			case reflect.String:
				fieldValue.SetString(value)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				intValue, err := strconv.Atoi(value)
				if err != nil {
					err := fmt.Errorf("invalid value for int name=%s, value=%s; example=%q", prefixedKey, value, example)
					errors = append(errors, err)
					continue
				}
				fieldValue.SetInt(int64(intValue))
			case reflect.Bool:
				boolValue, err := strconv.ParseBool(value)
				if err != nil {
					err := fmt.Errorf("invalid value for bool name=%s, value=%s; example=%q", prefixedKey, value, example)
					errors = append(errors, err)
					continue
				}
				fieldValue.SetBool(boolValue)
			case reflect.Slice:
				arrValue := strings.Split(value, SeparatorCharacter)
				// slice of what? if int type, then use atoi
				if fieldValue.Type().Elem().Kind() == reflect.Int {
					intSlice := make([]int, 0)
					for _, v := range arrValue {
						intValue, err := strconv.Atoi(v)
						if err != nil {
							err := fmt.Errorf("invalid value for int name=%s, value=%s; example=%q", prefixedKey, value, example)
							errors = append(errors, err)
							continue
						}
						intSlice = append(intSlice, intValue)
					}
					fieldValue.Set(reflect.ValueOf(intSlice))
				} else {
					fieldValue.Set(reflect.ValueOf(arrValue))
				}
			case reflect.Array: // array not supported
				err := fmt.Errorf("parsing into an array not supported: name=%s, value=%s; example=%q", prefixedKey, value, example)
				errors = append(errors, err)
				continue
			}
		}
	}

	if len(errors) > 0 {
		for _, err := range errors {
			logger.Println(err)
		}
		return ErrConfigInvalid
	}

	return nil
}
