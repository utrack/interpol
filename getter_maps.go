package interpol

import (
	"errors"
	"fmt"
	"reflect"
)

// getMapSelector retrieves the spawner based on the map's value type.
func getMapSelector(v reflect.Type) (getterFuncSpawner, error) {
	if v.Key().Kind() != reflect.String {
		return nil, errors.New("Non-string keyed maps are not supported!")
	}
	switch v.Elem().Kind() {
	case reflect.String:
		// simple string-string map
		return mapStringStringSpawner, nil
	case reflect.Slice:
		// map of slice
		switch v.Elem().Elem().Kind() {
		// safe to assume that's []byte
		case reflect.Uint8:
			return mapStringByteSpawner, nil
		}
	}

	// Check for Stringer interface
	if v.Elem().Implements(reflect.TypeOf((*fmt.Stringer)(nil)).Elem()) {
		return mapStringStringerSpawner, nil
	}

	// fallback to generic map spawner
	return mapStringInterfaceSpawner, nil
}

// mapStringString is used for map[string]string lookup.
func mapStringStringSpawner(v interface{}) (getterFunc, error) {
	m := v.(map[string]string)
	return func(key string) ([]byte, error) {
		value, ok := m[key]
		if !ok {
			return nil, ErrMapKeyNotFound
		}
		return []byte(value), nil
	}, nil
}

// mapStringStringer is used for map[string]fmt.Stringer lookup.
func mapStringStringerSpawner(v interface{}) (getterFunc, error) {
	refVal := reflect.ValueOf(v)
	return func(key string) ([]byte, error) {
		value := refVal.MapIndex(reflect.ValueOf(key))
		if !value.IsValid() {
			return nil, ErrMapKeyNotFound
		}
		return []byte(value.Interface().(fmt.Stringer).String()), nil
	}, nil
}

// mapStringByte is used for map[string][]byte lookups.
func mapStringByteSpawner(v interface{}) (getterFunc, error) {
	m := v.(map[string][]byte)
	return func(key string) ([]byte, error) {
		value, ok := m[key]
		if !ok {
			return nil, ErrMapKeyNotFound
		}
		return value, nil
	}, nil
}

func mapStringInterfaceSpawner(v interface{}) (getterFunc, error) {
	refVal := reflect.ValueOf(v)
	return func(key string) ([]byte, error) {
		value := refVal.MapIndex(reflect.ValueOf(key))
		if !value.IsValid() {
			return nil, ErrMapKeyNotFound
		}
		return getStringFromInterface(value.Interface())
	}, nil
}
