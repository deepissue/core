/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package utils

// Pagination 返回分页数据结构
type Pagination struct {
	Page       int `json:"page"`
	Size       int `json:"size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func NewPagination(total int64, size int64, page int) *Pagination {
	mod := total % size
	totalPages := total / size
	if 0 == totalPages {
		totalPages = 1
	}
	if mod > 0 && total > size {
		totalPages += 1
	}
	return &Pagination{Total: int(total), TotalPages: int(totalPages), Page: page, Size: int(size)}
}
