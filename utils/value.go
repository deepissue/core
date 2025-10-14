// Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func PtrFloat64(f float64) *float64 {
	return &f
}

func PtrFloat32(f float32) *float32 {
	return &f
}

func PtrString(f string) *string {
	return &f
}

func PtrInt(f int) *int {
	return &f
}

func PtrInt8(f int8) *int8 {
	return &f
}

func PtrInt16(f int16) *int16 {
	return &f
}

func PtrInt32(f int32) *int32 {
	return &f
}

func PtrInt64(f int64) *int64 {
	return &f
}

func PtrUint(f uint) *uint {
	return &f
}
func PtrUint8(f uint8) *uint8 {
	return &f
}

func PtrUint16(f uint16) *uint16 {
	return &f
}

func PtrUint32(f uint32) *uint32 {
	return &f
}

func PtrUint64(f uint64) *uint64 {
	return &f
}
func PtrBool(f bool) *bool {
	return &f
}

func PtrTime(f time.Time) *time.Time {
	return &f
}

func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

// IsEmpty 安全判断任意值是否为nil或零值
// 支持: pointer/slice/map/chan/func/interface + 结构体/字符串
func IsEmpty(v any) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)

	// Check if the value is invalid or nil for kinds that may be nil
	switch val.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		return val.IsNil()
	case reflect.String:
		return val.Len() == 0
	case reflect.Struct, reflect.Array:
		// Compare directly to zero value once
		zeroVal := reflect.Zero(val.Type())
		return reflect.DeepEqual(val.Interface(), zeroVal.Interface())
	case reflect.UnsafePointer:
		return val.Pointer() == 0
	default:
		zeroVal := reflect.Zero(val.Type()).Interface()
		return reflect.DeepEqual(v, zeroVal)
	}
}
