package toumin

import (
	"fmt"
	"strconv"
)

type FieldValue struct {
	value interface{}
}

func (v *FieldValue) Get() interface{} {
	return v.value
}

func (v *FieldValue) Addr() *interface{} {
	return &v.value
}

func (v *FieldValue) SetAddr(value interface{}) {
	switch s := value.(type) {
	case *string, *int, *int8, *int16, *int32, *int64, *[]byte:
		v.value = value
	case string, int, int8, int16, int32, int64, []byte:
		v.value = s
	}
}

func (v *FieldValue) Set(value interface{}) {
	switch value.(type) {
	case int:
		*v.value.(*int) = value.(int)
	case int8:
		*v.value.(*int8) = value.(int8)
	case int16:
		*v.value.(*int16) = value.(int16)
	case int32:
		*v.value.(*int32) = value.(int32)
	case int64:
		*v.value.(*int64) = value.(int64)
	case *int:
		*v.value.(*int) = *value.(*int)
	case *int8:
		*v.value.(*int8) = *value.(*int8)
	case *int16:
		*v.value.(*int16) = *value.(*int16)
	case *int32:
		*v.value.(*int32) = *value.(*int32)
	case *int64:
		*v.value.(*int64) = *value.(*int64)
	case string:
		x, ok := v.value.(*string)
		if ok {
			*x = value.(string)
		} else {
			v.value = value
		}
	case *string:
		*v.value.(*string) = *value.(*string)
	case []byte:
		*v.value.(*[]byte) = value.([]byte)
	case *[]byte:
		*v.value.(*[]byte) = *value.(*[]byte)
	}
}

func StrToInt(s string) int64 {
	if v, err := strconv.ParseInt(s, 0, 64); err == nil {
		return v
	}
	return 0
}

func (v *FieldValue) IsNil() bool {
	switch v.value.(type) {
	case nil:
		return true
	}
	return false
}

func (v *FieldValue) Int() int64 {
	switch value := v.value.(type) {
	case int:
		return int64(value)
	case int8:
		return int64(value)
	case int16:
		return int64(value)
	case int32:
		return int64(value)
	case int64:
		return value
	case *int:
		return int64(*value)
	case *int8:
		return int64(*value)
	case *int16:
		return int64(*value)
	case *int32:
		return int64(*value)
	case *int64:
		return *value
	case string:
		return StrToInt(value)
	case *string:
		return StrToInt(*value)
	case []byte:
		return StrToInt(string(value))
	case *[]byte:
		return StrToInt(string(*value))
	}
	return 0
}

func (v *FieldValue) String() string {
	switch value := v.value.(type) {
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", value)
	case *int:
		return fmt.Sprintf("%d", *value)
	case *int8:
		return fmt.Sprintf("%d", *value)
	case *int16:
		return fmt.Sprintf("%d", *value)
	case *int32:
		return fmt.Sprintf("%d", *value)
	case *int64:
		return fmt.Sprintf("%d", *value)
	case string:
		return value
	case *string:
		return *value
	case []byte:
		return string(value)
	case *[]byte:
		return string(*value)
	case nil:
		return "NULL"
	}
	return ""
}

type FieldData map[string]*FieldValue
