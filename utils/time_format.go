/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package utils

import "time"

type TimeFormat string

const (
	TimeFormatTime TimeFormat = "HH:mm:ss"
	TimeFormatDate TimeFormat = "yyyy-MM-dd"
	TimeFormatYear TimeFormat = "yyyy"
	TimeFormatFull TimeFormat = "yyyy-MM-dd HH:mm:ss"
)

// FormatTime
// 2006-01-02T15:04:05Z07:00
func FormatTime(o time.Time, fmt TimeFormat) string {
	switch fmt {
	case "yyyy-MM-dd":
		return o.Format("2006-01-02")
	case "yyyy-MM":
		return o.Format("2006-01")
	case "yyyy":
		return o.Format("2006")
	case "MM-dd":
		return o.Format("01-02")
	case "yyyy-MM-dd HH:mm:ss":
		return o.Format("2006-01-02 15:04:05")
	}
	return o.Format(time.RFC3339)
}
