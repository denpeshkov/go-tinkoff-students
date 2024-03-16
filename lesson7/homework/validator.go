package homework

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unsafe"
)

var ErrNotStruct = errors.New("wrong argument given, should be a struct")
var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
var ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")

type ValidationError struct {
	Err error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 1 {
		return v[0].Err.Error()
	}
	b := ([]byte)(v[0].Err.Error())
	for _, e := range v {
		b = append(b, "; "...)
		b = append(b, []byte(e.Err.Error())...)
	}
	// prevent allocation
	return unsafe.String(&b[0], len(b))
}

func Validate(v any) error {
	var errs ValidationErrors

	sv := reflect.ValueOf(v)
	st := sv.Type()
	if st.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	for i := 0; i < sv.NumField(); i++ {
		fT, fV := st.Field(i), sv.Field(i)
		tag, ok := fT.Tag.Lookup("validate")
		if ok && !fT.IsExported() {
			errs = append(errs, ValidationError{ErrValidateForUnexportedFields})
			continue
		} else if !ok {
			continue
		}

		val, err := getValidator(tag)
		if err != nil {
			errs = append(errs, ValidationError{err})
			continue
		}
		if err := val(fV.Interface()); err != nil {
			errs = append(errs, ValidationError{err})
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

func getValidator(tag string) (func(i any) error, error) {
	ss := strings.SplitN(tag, ":", 2)
	if len(ss) < 2 {
		return nil, ErrInvalidValidatorSyntax
	}

	switch ss[0] {
	case "len":
		return func(i any) error { return validateLen(i, ss[1]) }, nil
	case "in":
		return func(i any) error { return validateIn(i, ss[1]) }, nil
	case "min":
		return func(i any) error { return validateMin(i, ss[1]) }, nil
	case "max":
		return func(i any) error { return validateMax(i, ss[1]) }, nil
	}
	return nil, ErrInvalidValidatorSyntax
}

func validateLen(i any, s string) error {
	str, ok := i.(string)
	if !ok {
		return errors.New("field is not a string")
	}
	length, err := strconv.Atoi(s)
	if err != nil {
		return ErrInvalidValidatorSyntax
	}

	if len(str) != length {
		return errors.New("length of the field is incorrect")
	}
	return nil
}

func validateIn(i any, s string) error {
	strs := strings.Split(s, ",")
	if len(strs) == 0 || strs[0] == "" {
		return ErrInvalidValidatorSyntax
	}

	switch v := i.(type) {
	case int:
		nums := make([]int, len(strs))
		for i, s := range strs {
			v, err := strconv.Atoi(s)
			if err != nil {
				return ErrInvalidValidatorSyntax
			}
			nums[i] = v
		}
		if !slices.Contains(nums, v) {
			return fmt.Errorf("field value is not in the set: %v", nums)
		}
	case string:
		if !slices.Contains(strs, v) {
			return fmt.Errorf("field value is not in the set: %q", strs)
		}
	default:
		return errors.New("field is not a string nor an int")
	}
	return nil
}

func validateMin(i any, s string) error {
	minV, err := strconv.Atoi(s)
	if err != nil {
		return ErrInvalidValidatorSyntax
	}

	switch v := i.(type) {
	case int:
		if v < minV {
			return fmt.Errorf("field value is less than: %v", minV)
		}
	case string:
		if len(v) < minV {
			return fmt.Errorf("field length is less than: %v", minV)
		}
	default:
		return errors.New("field is not a string nor an int")
	}
	return nil
}

func validateMax(i any, s string) error {
	maxV, err := strconv.Atoi(s)
	if err != nil {
		return ErrInvalidValidatorSyntax
	}

	switch v := i.(type) {
	case int:
		if v > maxV {
			return fmt.Errorf("field value is greater than: %v", maxV)
		}
	case string:
		if len(v) > maxV {
			return fmt.Errorf("field length is greater than: %v", maxV)
		}
	default:
		return errors.New("field is not a string nor an int")
	}
	return nil
}
