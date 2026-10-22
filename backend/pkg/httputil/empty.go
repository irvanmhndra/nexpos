package httputil

import "reflect"

// emptyLists replaces nil slices in a response with empty ones, at the top
// level and inside structs and pointers, so lists always encode as [] and
// never as null. Clients can then iterate without null checks.
func emptyLists(data any) any {
	if data == nil {
		return nil
	}
	v := reflect.ValueOf(data)
	switch {
	case v.Kind() == reflect.Slice && v.IsNil():
		return reflect.MakeSlice(v.Type(), 0, 0).Interface()
	case v.Kind() == reflect.Struct:
		// A struct passed by value is not addressable: fill a copy.
		cp := reflect.New(v.Type()).Elem()
		cp.Set(v)
		fillLists(cp, 0)
		return cp.Interface()
	}
	fillLists(v, 0)
	return data
}

func fillLists(v reflect.Value, depth int) {
	if depth > 8 {
		return
	}
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			fillLists(v.Elem(), depth+1)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			fillLists(v.Index(i), depth+1)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if !v.Type().Field(i).IsExported() {
				continue
			}
			if f.Kind() == reflect.Slice && f.IsNil() && f.CanSet() {
				f.Set(reflect.MakeSlice(f.Type(), 0, 0))
				continue
			}
			fillLists(f, depth+1)
		}
	}
}
