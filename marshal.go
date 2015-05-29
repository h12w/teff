package teff

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

func Marshal(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte("nil"), nil
	}
	return marshal(reflectValue(v))
}

func marshal(v reflect.Value) ([]byte, error) {
	switch v.Type().Kind() {
	case reflect.Int:
		return []byte(fmt.Sprint(v.Interface())), nil
	case reflect.Slice:
		ss := [][]byte{}
		for i := 0; i < v.Len(); i++ {
			f, err := marshal(v.Index(i))
			if err != nil {
				return nil, err
			}
			ss = append(ss, f)
		}
		return bytes.Join(ss, []byte{'\n'}), nil
	}
	return nil, fmt.Errorf("marshal unsupported")
}

func Unmarshal(data []byte, v interface{}) error {
	if string(data) == "nil" {
		return nil
	}
	val := allocValue(v)
	switch val.Type().Kind() {
	case reflect.Int:
		i, err := strconv.Atoi(string(data))
		if err != nil {
			return err
		}
		val.SetInt(int64(i))
		return nil
	}
	return fmt.Errorf("unmarshal unsupported")
}

func reflectValue(value interface{}) reflect.Value {
	v := reflect.ValueOf(value)
	for v.Type().Kind() == reflect.Ptr && !v.IsNil() {
		v = reflect.Indirect(v)
	}
	return v
}

func allocValue(value interface{}) reflect.Value {
	v := reflect.ValueOf(value)
	for v.Type().Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = reflect.Indirect(v)
	}
	return v
}
