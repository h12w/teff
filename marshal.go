package teff

import (
	"fmt"
	"reflect"
	"strconv"
)

func Marshal(value interface{}) ([]byte, error) {
	if value == nil {
		return []byte("nil"), nil
	}
	v := reflectValue(value)
	switch v.Type().Kind() {
	case reflect.Slice:
	}
	return []byte(fmt.Sprint(v.Interface())), nil
}

func Unmarshal(data []byte, v interface{}) error {
	if string(data) == "nil" {
		return nil
	}
	val := reflectValue(v)
	switch val.Type().Kind() {
	case reflect.Int:
		i, err := strconv.Atoi(string(data))
		if err != nil {
			return err
		}
		val.SetInt(int64(i))
		return nil
	}
	return fmt.Errorf("unsupported")
}

func reflectValue(value interface{}) reflect.Value {
	v := reflect.ValueOf(value)
	for v.Type().Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = reflect.Indirect(v)
	}
	return v
}
