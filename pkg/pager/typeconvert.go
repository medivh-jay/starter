package pager

import (
	"strconv"
	"strings"
)

// StringToInt convert string to int
//  实际上, 如果传入的并不是string的话这里不会反悔int，而是返回传入的原始数据
func StringToInt(val interface{}) interface{} {
	v, ok := val.(string)
	if !ok {
		return val
	}
	rs, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return rs
}

// StringToFloat32 convert string to float32
//  实际上, 如果传入的并不是string的话这里不会反悔int，而是返回传入的原始数据
func StringToFloat32(val interface{}) interface{} {
	v, ok := val.(string)
	if !ok {
		return val
	}
	rs, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return 0
	}
	return rs
}

// StringToFloat64 convert string to float64
//  实际上, 如果传入的并不是string的话这里不会反悔int，而是返回传入的原始数据
func StringToFloat64(val interface{}) interface{} {
	v, ok := val.(string)
	if !ok {
		return val
	}
	rs, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return rs
}

// StringToBool convert string to bool
func StringToBool(val interface{}) interface{} {
	if strings.ToLower(val.(string)) == "true" || val == "1" || (val != "0" && len(val.(string)) > 0) {
		return true
	}
	return false
}
