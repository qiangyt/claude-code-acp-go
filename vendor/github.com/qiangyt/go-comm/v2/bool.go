package comm

import (
	"fmt"
	"reflect"
	"strings"
)

// RequiredBoolP returns the bool value of the key in the map. If either parsing error or the key is not
// found, raise a panic.
func RequiredBoolP(hint string, key string, m map[string]any) bool {
	r, err := RequiredBool(hint, key, m)
	if err != nil {
		panic(err)
	}
	return r
}

// RequiredBool returns the bool value of the key in the map. If either parsing error or the key is not
// found, an error is returned.
func RequiredBool(hint string, key string, m map[string]any) (bool, error) {
	v, has := m[key]
	if !has {
		return false, LocalizeError("error.required", map[string]interface{}{
			"Hint": hint,
			"Key":  key,
		})
	}

	return Bool(hint+"."+key, v)
}

// OptionalBoolP returns the bool value of the key in the map. If parsing error occurred,
// raise a panic. If the key is not found, return the default value.
func OptionalBoolP(hint string, key string, m map[string]any, devault bool) (result bool, has bool) {
	var err error
	result, has, err = OptionalBool(hint, key, m, devault)
	if err != nil {
		panic(err)
	}
	return result, has
}

// OptionalBool returns the bool value of the key in the map. If parsing error occrred,
// returns the error. If the key is not found, return the default value.
func OptionalBool(hint string, key string, m map[string]any, devault bool) (result bool, has bool, err error) {
	var v any
	v, has = m[key]
	if !has {
		result = devault
		return result, has, err
	}

	result, err = Bool(hint+"."+key, v)
	return result, has, err
}

// Cast the value to bool. If parsing error occurred, raise a panic.
func BoolP(hint string, v any) bool {
	r, err := Bool(hint, v)
	if err != nil {
		panic(err)
	}
	return r
}

// Cast the value to bool. If parsing error occurred, returns the error.
func Bool(hint string, v any) (bool, error) {
	r, ok := v.(bool)
	if !ok {
		if str, isStr := v.(string); isStr {
			str = strings.ToLower(strings.TrimSpace(str))
			if str == "true" {
				return true, nil
			}
			if str == "false" {
				return false, nil
			}
		}
		return false, LocalizeError("error.type.bool", map[string]interface{}{
			"Hint":  hint,
			"Type":  reflect.TypeOf(v),
			"Value": v,
		})
	}
	return r, nil
}

func BoolArrayP(hint string, v any) []bool {
	r, err := BoolArray(hint, v)
	if err != nil {
		panic(err)
	}
	return r
}

func BoolArray(hint string, v any) ([]bool, error) {
	r, ok := v.([]bool)
	if !ok {
		if r1, ok1 := v.([]any); ok1 {
			var err error
			r = make([]bool, len(r1))
			for i, v := range r1 {
				if r[i], err = Bool(fmt.Sprintf("%s[%d]", hint, i), v); err != nil {
					break
				}
			}
			if err == nil {
				return r, nil
			}
		} else if r0, err := Bool(hint, v); err == nil {
			return []bool{r0}, nil
		}
		return nil, LocalizeError("error.type.bool_array", map[string]interface{}{
			"Hint":  hint,
			"Type":  reflect.TypeOf(v),
			"Value": v,
		})
	}
	return r, nil
}

func BoolMapP(hint string, v any) map[string]bool {
	r, err := BoolMap(hint, v)
	if err != nil {
		panic(err)
	}
	return r
}

func BoolMap(hint string, v any) (map[string]bool, error) {
	r, ok := v.(map[string]bool)
	if !ok {
		if r1, ok1 := v.(map[string]any); ok1 {
			var err error
			r = map[string]bool{}
			for k, v := range r1 {
				if r[k], err = Bool(hint+"."+k, v); err != nil {
					break
				}
			}
			if err == nil {
				return r, nil
			}
		} else if r0, ok0 := v.(string); ok0 {
			if posOfColon := strings.Index(r0, ":"); posOfColon > 0 && posOfColon != len(r0)-1 {
				if value, err := Bool(hint, strings.TrimSpace(r0[posOfColon+1:])); err == nil {
					key := strings.TrimSpace(r0[:posOfColon])
					return map[string]bool{key: value}, nil
				}
			}
		}
		return nil, LocalizeError("error.type.bool_map", map[string]interface{}{
			"Hint":  hint,
			"Type":  reflect.TypeOf(v),
			"Value": v,
		})
	}
	return r, nil
}
