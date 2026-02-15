package comm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// RequiredFloatP returns the float64 value of the key in the map. If either parsing error or the key is not
// found, raise a panic.
func RequiredFloatP(hint string, key string, m map[string]any) float64 {
	r, err := RequiredFloat(hint, key, m)
	if err != nil {
		panic(err)
	}
	return r
}

// RequiredFloat returns the float64 value of the key in the map. If either parsing error or the key is not
// found, an error is returned.
func RequiredFloat(hint string, key string, m map[string]any) (float64, error) {
	v, has := m[key]
	if !has {
		return 0, LocalizeError("error.required", map[string]interface{}{
			"Hint": hint,
			"Key":  key,
		})
	}

	return Float(hint+"."+key, v)
}

// OptionalFloatP returns the float64 value of the key in the map. If parsing error occurred,
// raise a panic. If the key is not found, return the default value.
func OptionalFloatP(hint string, key string, m map[string]any, devault float64) (result float64, has bool) {
	var err error
	result, has, err = OptionalFloat(hint, key, m, devault)
	if err != nil {
		panic(err)
	}
	return result, has
}

// OptionalFloat returns the float64 value of the key in the map. If parsing error occrred,
// returns the error. If the key is not found, return the default value.
func OptionalFloat(hint string, key string, m map[string]any, devault float64) (result float64, has bool, err error) {
	var v any

	v, has = m[key]
	if !has {
		result = devault
		return result, has, err
	}

	result, err = Float(hint+"."+key, v)
	return result, has, err
}

// Cast the value to float64. If parsing error occurred, raise a panic.
func FloatP(hint string, v any) float64 {
	r, err := Float(hint, v)
	if err != nil {
		panic(err)
	}
	return r
}

// Cast the value to float64. If parsing error occurred, returns the error.
func Float(hint string, v any) (float64, error) {
	r, ok := v.(float64)
	if !ok {
		if i, isInt := v.(int); isInt {
			return float64(i), nil
		}
		if i32, isInt32 := v.(int32); isInt32 {
			return float64(i32), nil
		}
		if i64, isInt64 := v.(int64); isInt64 {
			return float64(i64), nil
		}
		if s, isString := v.(string); isString {
			if parsed, err := strconv.ParseFloat(s, 64); err == nil {
				return parsed, nil
			}
		}
		return 0, LocalizeError("error.type.float", map[string]interface{}{
			"Hint":  hint,
			"Type":  reflect.TypeOf(v),
			"Value": v,
		})
	}
	return r, nil
}

func FloatArrayP(hint string, v any) []float64 {
	r, err := FloatArray(hint, v)
	if err != nil {
		panic(err)
	}
	return r
}

func FloatArray(hint string, v any) ([]float64, error) {
	r, ok := v.([]float64)
	if !ok {
		if r1, ok1 := v.([]any); ok1 {
			var err error
			r = make([]float64, len(r1))
			for i, v := range r1 {
				if r[i], err = Float(fmt.Sprintf("%s[%d]", hint, i), v); err != nil {
					break
				}
			}
			if err == nil {
				return r, nil
			}
		} else if r0, err := Float(hint, v); err == nil {
			return []float64{r0}, nil
		}
		return nil, LocalizeError("error.type.float_array", map[string]interface{}{
			"Hint":  hint,
			"Type":  reflect.TypeOf(v),
			"Value": v,
		})
	}
	return r, nil
}

func FloatMapP(hint string, v any) map[string]float64 {
	r, err := FloatMap(hint, v)
	if err != nil {
		panic(err)
	}
	return r
}

func FloatMap(hint string, v any) (map[string]float64, error) {
	r, ok := v.(map[string]float64)
	if !ok {
		if r1, ok1 := v.(map[string]any); ok1 {
			var err error
			r = map[string]float64{}
			for k, v := range r1 {
				if r[k], err = Float(hint+"."+k, v); err != nil {
					break
				}
			}
			if err == nil {
				return r, nil
			}
		} else if r0, ok0 := v.(string); ok0 {
			if posOfColon := strings.Index(r0, ":"); posOfColon > 0 && posOfColon != len(r0)-1 {
				if value, err := Float(hint, strings.TrimSpace(r0[posOfColon+1:])); err == nil {
					key := strings.TrimSpace(r0[:posOfColon])
					return map[string]float64{key: value}, nil
				}
			}
		}
		return nil, LocalizeError("error.type.float_map", map[string]interface{}{
			"Hint":  hint,
			"Type":  reflect.TypeOf(v),
			"Value": v,
		})
	}
	return r, nil
}
