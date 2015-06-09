package teff

import (
	"bytes"
	"fmt"
	"h12.me/teff/core"
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
	list, err := core.Parse(bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	val := allocValue(v)
	switch val.Type().Kind() {
	case reflect.Int:
		return unmarshalNode(list[0], val)
	case reflect.Slice:
		for i, node := range list {
			val.Set(reflect.Append(val, reflect.New(val.Type().Elem()).Elem()))
			elem := val.Index(i)
			if err := unmarshalNode(node, elem); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("unmarshal unsupported")
}

func unmarshalNode(node core.Node, v reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.Int:
		i, err := strconv.Atoi(node.Value)
		if err != nil {
			return err
		}
		v.SetInt(int64(i))
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
