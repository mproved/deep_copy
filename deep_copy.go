package deep_copy

import (
	"fmt"
	"reflect"
)

var copiers map[reflect.Kind]func(any, map[uintptr]any) (any, error)

func init() {
	copiers = map[reflect.Kind]func(any, map[uintptr]any) (any, error){
		reflect.Bool:       copyPrimitive,
		reflect.Int:        copyPrimitive,
		reflect.Int8:       copyPrimitive,
		reflect.Int16:      copyPrimitive,
		reflect.Int32:      copyPrimitive,
		reflect.Int64:      copyPrimitive,
		reflect.Uint:       copyPrimitive,
		reflect.Uint8:      copyPrimitive,
		reflect.Uint16:     copyPrimitive,
		reflect.Uint32:     copyPrimitive,
		reflect.Uint64:     copyPrimitive,
		reflect.Uintptr:    copyPrimitive,
		reflect.Float32:    copyPrimitive,
		reflect.Float64:    copyPrimitive,
		reflect.Complex64:  copyPrimitive,
		reflect.Complex128: copyPrimitive,
		reflect.String:     copyPrimitive,
		reflect.Array:      copyArray,
		reflect.Slice:      copySlice,
		reflect.Map:        copyMap,
		reflect.Ptr:        copyPointer,
		reflect.Struct:     copyStruct,
	}
}

func MustCopy(x any) any {
	copied, err := Copy(x)

	if err != nil {
		panic(err)
	}

	return copied
}

func Copy(value any) (any, error) {
	pointers := make(map[uintptr]any)

	return copyWrapped(value, pointers)
}

func copyWrapped(value any, pointers map[uintptr]any) (any, error) {
	valueOf := reflect.ValueOf(value)

	if !valueOf.IsValid() {
		return value, nil
	}

	if copier, ok := copiers[valueOf.Kind()]; ok {
		return copier(value, pointers)
	}

	typeOf := reflect.TypeOf(value)

	return nil, fmt.Errorf("unable to make a deep copy of %v (type: %v) - kind %v is not supported", value, typeOf, valueOf.Kind())
}

func copyPrimitive(value any, pointers map[uintptr]any) (any, error) {
	kind := reflect.ValueOf(value).Kind()

	if kind == reflect.Array ||
		kind == reflect.Chan ||
		kind == reflect.Func ||
		kind == reflect.Interface ||
		kind == reflect.Map ||
		kind == reflect.Ptr ||
		kind == reflect.Slice ||
		kind == reflect.Struct ||
		kind == reflect.UnsafePointer {

		return nil, fmt.Errorf("unable to copy %v (a %v) as a primitive", value, kind)
	}

	return value, nil
}

func copyArray(value any, pointers map[uintptr]any) (any, error) {
	valueOf := reflect.ValueOf(value)

	if valueOf.Kind() != reflect.Array {
		return nil, fmt.Errorf("must pass a value with kind of Array; got %v", valueOf.Kind())
	}

	typeOf := reflect.TypeOf(value)
	size := typeOf.Len()
	copied := reflect.New(reflect.ArrayOf(size, typeOf.Elem())).Elem()

	for i := 0; i < size; i++ {
		item, err := copyWrapped(valueOf.Index(i).Interface(), pointers)

		if err != nil {
			return nil, fmt.Errorf("failed to clone array item at index %v: %v", i, err)
		}

		copied.Index(i).Set(reflect.ValueOf(item))
	}

	return copied.Interface(), nil
}

func copySlice(value any, pointers map[uintptr]any) (any, error) {
	valueOf := reflect.ValueOf(value)

	if valueOf.Kind() != reflect.Slice {
		return nil, fmt.Errorf("must pass a value with kind of Slice; got %v", valueOf.Kind())
	}

	size := valueOf.Len()
	typeOf := reflect.TypeOf(value)
	copied := reflect.MakeSlice(typeOf, size, size)

	for i := 0; i < size; i++ {
		item, err := copyWrapped(valueOf.Index(i).Interface(), pointers)

		if err != nil {
			return nil, fmt.Errorf("failed to clone slice item at index %v: %v", i, err)
		}

		itemValueOf := reflect.ValueOf(item)

		if itemValueOf.IsValid() {
			copied.Index(i).Set(itemValueOf)
		}
	}

	return copied.Interface(), nil
}

func copyMap(value any, pointers map[uintptr]any) (any, error) {
	valueOf := reflect.ValueOf(value)

	if valueOf.Kind() != reflect.Map {
		return nil, fmt.Errorf("must pass a value with kind of Map; got %v", valueOf.Kind())
	}

	typeOf := reflect.TypeOf(value)
	copied := reflect.MakeMapWithSize(typeOf, valueOf.Len())
	iter := valueOf.MapRange()

	for iter.Next() {
		mapKey, err := copyWrapped(iter.Key().Interface(), pointers)

		if err != nil {
			return nil, fmt.Errorf("failed to clone the map key %v: %v", iter.Key().Interface(), err)
		}

		mapValue, err := copyWrapped(iter.Value().Interface(), pointers)

		if err != nil {
			return nil, fmt.Errorf("failed to clone map value %v: %v", iter.Value().Interface(), err)
		}

		copied.SetMapIndex(reflect.ValueOf(mapKey), reflect.ValueOf(mapValue))
	}

	return copied.Interface(), nil
}

func copyPointer(value any, pointers map[uintptr]any) (any, error) {
	valueOf := reflect.ValueOf(value)

	if valueOf.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("must pass a value with kind of Ptr; got %v", valueOf.Kind())
	}

	if valueOf.IsNil() {
		typeOf := reflect.TypeOf(value)
		return reflect.Zero(typeOf).Interface(), nil
	}

	address := valueOf.Pointer()

	if copied, ok := pointers[address]; ok {
		return copied, nil
	}

	typeOf := reflect.TypeOf(value)

	copied := reflect.New(typeOf.Elem())

	pointers[address] = copied.Interface()

	item, err := copyWrapped(valueOf.Elem().Interface(), pointers)

	if err != nil {
		return nil, fmt.Errorf("failed to copy the value under the pointer %v: %v", valueOf, err)
	}

	itemValueOf := reflect.ValueOf(item)

	if itemValueOf.IsValid() {
		copied.Elem().Set(reflect.ValueOf(item))
	}

	return copied.Interface(), nil
}

func copyStruct(value any, pointers map[uintptr]any) (any, error) {
	valueOf := reflect.ValueOf(value)

	if valueOf.Kind() != reflect.Struct {
		return nil, fmt.Errorf("must pass a value with kind of Struct; got %v", valueOf.Kind())
	}

	typeOf := reflect.TypeOf(value)
	copied := reflect.New(typeOf)

	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		fieldValue := valueOf.Field(i)

		if !field.IsExported() {
			continue
		}

		item, err := copyWrapped(fieldValue.Interface(), pointers)

		if err != nil {
			return nil, fmt.Errorf("failed to copy the field %v in the struct %#v: %v", typeOf.Field(i).Name, value, err)
		}

		copied.Elem().Field(i).Set(reflect.ValueOf(item))
	}

	return copied.Elem().Interface(), nil
}
