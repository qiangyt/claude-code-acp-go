package comm

import (
	"fmt"
	"reflect"
)

func RequiredMapP(hint string, key string, m map[string]any) map[string]any {
	r, err := RequiredMap(hint, key, m)
	if err != nil {
		panic(err)
	}
	return r
}

func RequiredMap(hint string, key string, m map[string]any) (map[string]any, error) {
	v, has := m[key]
	if !has {
		return nil, fmt.Errorf("%s.%s is required", hint, key)
	}

	return Map(hint+"."+key, v)
}

func OptionalMapP(hint string, key string, m map[string]any, defaultresult map[string]any) (result map[string]any, has bool) {
	var err error
	result, has, err = OptionalMap(hint, key, m, defaultresult)
	if err != nil {
		panic(err)
	}
	return result, has
}

func OptionalMap(hint string, key string, m map[string]any, defaultresult map[string]any) (result map[string]any, has bool, err error) {
	var v any
	v, has = m[key]
	if !has {
		result = defaultresult
		return result, has, err
	}

	result, err = Map(hint+"."+key, v)
	return result, has, err
}

func MapP(hint string, v any) map[string]any {
	r, err := Map(hint, v)
	if err != nil {
		panic(err)
	}
	return r
}

func Map(hint string, v any) (map[string]any, error) {
	r, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be a map[string]any, but now it is a %v(%v)", hint, reflect.TypeOf(v), v)
	}
	return r, nil
}

func RequiredMapArrayP(hint string, key string, m map[string]any) []map[string]any {
	r, err := RequiredMapArray(hint, key, m)
	if err != nil {
		panic(err)
	}
	return r
}

func RequiredMapArray(hint string, key string, m map[string]any) ([]map[string]any, error) {
	v, has := m[key]
	if !has {
		return nil, fmt.Errorf("%s.%s is required", hint, key)
	}

	return MapArray(hint+"."+key, v)
}

func OptionalMapArrayP(hint string, key string, m map[string]any, defaultresult []map[string]any) (result []map[string]any, has bool) {
	var err error
	result, has, err = OptionalMapArray(hint, key, m, defaultresult)
	if err != nil {
		panic(err)
	}
	return result, has
}

func OptionalMapArray(hint string, key string, m map[string]any, defaultresult []map[string]any) (result []map[string]any, has bool, err error) {
	var v any
	v, has = m[key]
	if !has {
		result = defaultresult
		return result, has, err
	}

	result, err = MapArray(hint+"."+key, v)
	return result, has, err
}

func MapArrayP(hint string, v any) []map[string]any {
	r, err := MapArray(hint, v)
	if err != nil {
		panic(err)
	}
	return r
}

// MapArray casts the input any value to []map[string]any if the input value type matches,
// otherwise, return error.
func MapArray(hint string, v any) ([]map[string]any, error) {
	r, ok := v.([]map[string]any)
	if !ok {
		if r1, ok1 := v.([]any); ok1 {
			var err error
			r = make([]map[string]any, len(r1))
			for i, v := range r1 {
				if r[i], err = Map(fmt.Sprintf("%s[%d]", hint, i), v); err != nil {
					break
				}
			}
			if err == nil {
				return r, nil
			}
		} else if r0, ok0 := v.(map[string]any); ok0 {
			return []map[string]any{r0}, nil
		}
		return nil, fmt.Errorf("%s must be a map[string]any array, but now it is a %v(%v)", hint, reflect.TypeOf(v), v)
	}
	return r, nil
}
