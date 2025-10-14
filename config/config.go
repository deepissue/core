/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package config

type Config struct {
	Extras Params `json:"extras" hcl:"extras,block"`
}

type Params map[string]Param

func (e Params) GetExtra(key string) Param {
	return e[key]
}

type Param map[string]any

func (e Param) GetValue(key string) any {
	return e[key]
}
func (e Param) GetString(key string) string {
	if val, ok := e[key]; ok {
		return val.(string)
	}
	return ""
}

func (e Param) GetInt(key string) (int, bool) {
	if val, ok := e[key]; ok {
		return val.(int), true
	}
	return 0, false
}

func (e Param) GetInt64(key string) (int64, bool) {
	if val, ok := e[key]; ok {
		return val.(int64), true
	}
	return 0, false
}
func (e Param) GetUInt(key string) (uint, bool) {
	if val, ok := e[key]; ok {
		return val.(uint), true
	}
	return 0, false
}

func (e Param) GetUInt64(key string) (uint64, bool) {
	if val, ok := e[key]; ok {
		return val.(uint64), true
	}
	return 0, false
}

func (e Param) GetParam(key string) Param {
	if val, ok := e[key]; ok {
		params := val.([]map[string]any)
		if len(params) > 0 {
			return params[0]
		}
	}
	return nil
}

func (e Param) Keys() []string {
	var keys []string
	for k, _ := range e {
		keys = append(keys, k)
	}
	return keys
}
