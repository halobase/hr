package hr

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

func bind(v interface{}, values map[string][]string, tag string) error {
	if v == nil {
		return nil
	}

	typ := reflect.TypeOf(v).Elem()
	val := reflect.ValueOf(v).Elem()

	if typ.Kind() != reflect.Struct {
		return errors.New("binding element must be a struct")
	}

	for i := 0; i < typ.NumField(); i++ {
		fieldTyp := typ.Field(i)
		fieldVal := val.Field(i)

		if fieldTyp.Anonymous {
			if val.Kind() == reflect.Ptr {
				fieldVal = fieldVal.Elem()
			}
		}
		if !fieldVal.CanSet() {
			// ignore private fields.
			continue
		}

		kind := fieldVal.Kind()
		key := fieldTyp.Tag.Get(tag)

		if len(key) == 0 {
			// ignore fields untagged.
			continue
		}

		opt := key[len(key)-1] == '?'
		if opt {
			key = key[:len(key)-1]
		}

		vals, ok := values[key]
		if !ok {
			if !opt {
				return fmt.Errorf("missing field: %s", key)
			}
			continue
		}

		// we have to check if it is a self-defined type here.
		// because types like net.IP (alias of []byte) is taken
		// as of type slice by Go.
		if iface, ok := fieldVal.Addr().Interface().(Type); ok {
			if err := iface.Parse(vals[0]); err != nil {
				return err
			}
			continue
		}

		nvals := len(vals)
		if kind == reflect.Slice && nvals > 0 {
			sliceOf := fieldVal.Type().Elem().Kind()
			slice := reflect.MakeSlice(fieldVal.Type(), nvals, nvals)
			for j := 0; j < nvals; j++ {
				if err := set(sliceOf, vals[j], slice.Index(j)); err != nil {
					return err
				}
			}
			fieldVal.Set(slice)
			continue
		}
		if err := set(kind, vals[0], fieldVal); err != nil {
			return err
		}
	}
	return nil
}

func set(k reflect.Kind, s string, v reflect.Value) error {
	switch k {
	case reflect.Pointer:
		return setPtr(s, v)
	case reflect.Int:
		return setInt(s, 0, v)
	case reflect.Int8:
		return setInt(s, 8, v)
	case reflect.Int16:
		return setInt(s, 16, v)
	case reflect.Int32:
		return setInt(s, 32, v)
	case reflect.Int64:
		return setInt(s, 64, v)
	case reflect.Uint:
		return setUint(s, 0, v)
	case reflect.Uint8:
		return setUint(s, 8, v)
	case reflect.Uint16:
		return setUint(s, 16, v)
	case reflect.Uint32:
		return setUint(s, 32, v)
	case reflect.Uint64:
		return setUint(s, 64, v)
	case reflect.Bool:
		return setBool(s, v)
	case reflect.Float32:
		return setFloat(s, 32, v)
	case reflect.Float64:
		return setFloat(s, 64, v)
	case reflect.String:
		v.SetString(s)
	default:
		return setCustom(s, v)
	}
	return nil
}

func setInt(s string, bits int, v reflect.Value) error {
	if len(s) == 0 {
		s = "0"
	}
	i64, err := strconv.ParseInt(s, 10, bits)
	if err == nil {
		v.SetInt(i64)
	}
	return err
}

func setUint(s string, bits int, v reflect.Value) error {
	if len(s) == 0 {
		s = "0"
	}
	u64, err := strconv.ParseUint(s, 10, bits)
	if err == nil {
		v.SetUint(u64)
	}
	return err
}

func setFloat(s string, bits int, v reflect.Value) error {
	if len(s) == 0 {
		s = "0.0"
	}
	f64, err := strconv.ParseFloat(s, bits)
	if err == nil {
		v.SetFloat(f64)
	}
	return err
}

func setBool(s string, v reflect.Value) error {
	if len(s) == 0 {
		s = "false"
	}
	b, err := strconv.ParseBool(s)
	if err == nil {
		v.SetBool(b)
	}
	return err
}

func setPtr(s string, v reflect.Value) error {
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	return set(v.Elem().Kind(), s, v.Elem())
}

func setCustom(s string, v reflect.Value) error {
	iface, ok := v.Addr().Interface().(Type)
	if !ok {
		return errors.New("bad type")
	}
	return iface.Parse(s)
}
