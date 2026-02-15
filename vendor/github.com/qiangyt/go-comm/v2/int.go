package comm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// RequiredIntP returns the int value of the key in the map. If either parsing error or the key is not
// found, raise a panic.
func RequiredIntP(hint string, key string, m map[string]any) int {
	r, err := RequiredInt(hint, key, m)
	if err != nil {
		panic(err)
	}
	return r
}

// RequiredInt returns the int value of the key in the map. If either parsing error or the key is not
// found, an error is returned.
func RequiredInt(hint string, key string, m map[string]any) (int, error) {
	v, has := m[key]
	if !has {
		return 0, LocalizeError("error.required", map[string]interface{}{
			"Hint": hint,
			"Key":  key,
		})
	}

	return Int(hint+"."+key, v)
}

// OptionalIntP returns the int value of the key in the map. If parsing error occurred,
// raise a panic. If the key is not found, return the default value.
func OptionalIntP(hint string, key string, m map[string]any, devault int) (result int, has bool) {
	var err error
	result, has, err = OptionalInt(hint, key, m, devault)
	if err != nil {
		panic(err)
	}
	return result, has
}

// OptionalInt returns the int value of the key in the map. If parsing error occrred,
// returns the error. If the key is not found, return the default value.
func OptionalInt(hint string, key string, m map[string]any, devault int) (result int, has bool, err error) {
	var v any

	v, has = m[key]
	if !has {
		result = devault
		return result, has, err
	}

	result, err = Int(hint+"."+key, v)
	return result, has, err
}

// Cast the value to int. If parsing error occurred, raise a panic.
func IntP(hint string, v any) int {
	r, err := Int(hint, v)
	if err != nil {
		panic(err)
	}
	return r
}

// Cast the value to int. If parsing error occurred, returns the error.
func Int(hint string, v any) (int, error) {
	r, ok := v.(int)
	if !ok {
		if f32, isFloat32 := v.(float32); isFloat32 {
			return int(f32), nil
		}
		if f64, isFloat64 := v.(float64); isFloat64 {
			return int(f64), nil
		}
		if s, isString := v.(string); isString {
			if parsed, err := strconv.Atoi(s); err == nil {
				return parsed, nil
			}
		}
		return 0, LocalizeError("error.type.int", map[string]interface{}{
			"Hint":  hint,
			"Type":  reflect.TypeOf(v),
			"Value": v,
		})
	}
	return r, nil
}

func IntArrayP(hint string, v any) []int {
	r, err := IntArray(hint, v)
	if err != nil {
		panic(err)
	}
	return r
}

func IntArray(hint string, v any) ([]int, error) {
	r, ok := v.([]int)
	if !ok {
		if r1, ok1 := v.([]any); ok1 {
			var err error
			r = make([]int, len(r1))
			for i, v := range r1 {
				if r[i], err = Int(fmt.Sprintf("%s[%d]", hint, i), v); err != nil {
					break
				}
			}
			if err == nil {
				return r, nil
			}
		} else if r0, err := Int(hint, v); err == nil {
			return []int{r0}, nil
		}
		return nil, LocalizeError("error.type.int_array", map[string]interface{}{
			"Hint":  hint,
			"Type":  reflect.TypeOf(v),
			"Value": v,
		})
	}
	return r, nil
}

func IntMapP(hint string, v any) map[string]int {
	r, err := IntMap(hint, v)
	if err != nil {
		panic(err)
	}
	return r
}

func IntMap(hint string, v any) (map[string]int, error) {
	r, ok := v.(map[string]int)
	if !ok {
		if r1, ok1 := v.(map[string]any); ok1 {
			var err error
			r = map[string]int{}
			for k, v := range r1 {
				if r[k], err = Int(hint+"."+k, v); err != nil {
					break
				}
			}
			if err == nil {
				return r, nil
			}
		} else if r0, ok0 := v.(string); ok0 {
			if posOfColon := strings.Index(r0, ":"); posOfColon > 0 && posOfColon != len(r0)-1 {
				if value, err := Int(hint, strings.TrimSpace(r0[posOfColon+1:])); err == nil {
					key := strings.TrimSpace(r0[:posOfColon])
					return map[string]int{key: value}, nil
				}
			}
		}
		return nil, LocalizeError("error.type.int_map", map[string]interface{}{
			"Hint":  hint,
			"Type":  reflect.TypeOf(v),
			"Value": v,
		})
	}
	return r, nil
}
