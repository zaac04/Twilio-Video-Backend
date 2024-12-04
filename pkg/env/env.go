package env

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Field struct {
	Key   string
	Value string
	Type  string
}

func Load_env(fl string, env interface{}) (err error) {
	err = godotenv.Load(fl)
	if err != nil {
		fmt.Println(err)
	}
	return verify_env(env)
}

func verify_env(s interface{}) error {
	StructVal := reflect.ValueOf(s)

	if StructVal.Kind() != reflect.Ptr || StructVal.Elem().Kind() != reflect.Struct {
		return errors.New("pointer expected recived struct")
	}
	Struct := StructVal.Elem()
	structType := Struct.Type()

	for i := 0; i < Struct.NumField(); i++ {
		var field Field
		err := getFieldMeta(structType, &field, i)
		if err != nil {
			return fmt.Errorf("%s", err.Error())
		}

		curr_field := Struct.Field(i)
		if !curr_field.CanSet() {
			return fmt.Errorf("field %s cannot be set", field.Key)
		}

		err = validate(&field)
		if err != nil {
			return fmt.Errorf("validation failed for field:%v, check env value provided", field.Value)
		}

		converted_val := convertToDataType(field.Value, curr_field.Type())
		val := reflect.ValueOf(converted_val)

		if val.Type().ConvertibleTo(curr_field.Type()) {
			curr_field.Set(val.Convert(curr_field.Type()))
		} else {
			return fmt.Errorf("cannot set field %s with value of type %T", field.Key, curr_field.Type())
		}
	}
	return nil
}

func getFieldMeta(valType reflect.Type, field *Field, i int) error {
	var env_name string

	tags := valType.Field(i).Tag.Get("env")
	if tags == "" {
		return fmt.Errorf("tagging the struct with 'env' is must")
	}

	tagPair := splitTags(tags)
	if tagPair == nil {
		return fmt.Errorf("failed to split tags into pairs")
	}

	if val, ok := tagPair["name"]; ok {
		env_name = val
	} else {
		env_name = valType.Field(i).Name
	}

	if val, ok := tagPair["type"]; ok {
		field.Type = val
	} else {
		field.Type = "string"
	}

	field.Key = valType.Field(i).Name
	field.Value = os.Getenv(env_name)

	return nil
}

func convertToDataType(value any, TargetType reflect.Type) interface{} {
	if TargetType.Kind() == reflect.Int {
		uintValue, _ := strconv.ParseInt(value.(string), 10, 32)
		return int(uintValue)
	}

	if TargetType.Kind() == reflect.String {
		return value.(string)
	}
	return value
}

func splitTags(s string) (data map[string]string) {
	data = make(map[string]string)
	pairs := strings.Split(s, ",")
	for _, p := range pairs {
		val := strings.Split(p, ":")
		if len(val) >= 2 {
			data[val[0]] = val[1]
		}
	}
	return data
}

func validate(field *Field) error {
	switch field.Type {
	case "url":
		parsedURL, err := url.Parse(field.Value)
		if err != nil {
			return fmt.Errorf("Invalid URL: %q, Error: %v\n", field.Value, err)
		}
		if parsedURL.Scheme == "" || parsedURL.Host == "" {
			return fmt.Errorf("Invalid URL: %q, Missing scheme or host\n", field.Value)
		}
		return nil
	default:
		if field.Value != "" {
			return nil
		} else {
			return nil
		}
	}
}
