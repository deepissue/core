/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package server

import (
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/deepissue/core/authorities"
	"github.com/deepissue/core/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			return jsonTag
		}
		return field.Name
	})
}

type Context struct {
	*gin.Context
	Authorized *authorities.Authorized
	RemoteAddr string
	ClientID   string
	Header     http.Header
}

func NewContext(c *gin.Context) *Context {
	ctx := &Context{
		Context:    c,
		RemoteAddr: utils.GetRemoteAddr(c.Request),
		ClientID:   c.GetHeader(ClientIDKey),
		Header:     c.Request.Header,
	}

	return ctx
}

func (c *Context) ValidateStruct(out any) error {
	return validate.Struct(out)
}

func (c *Context) PageNumber() int {

	page := c.Context.DefaultQuery("page", "1")
	ret, err := strconv.Atoi(page)
	if err != nil {
		return 1
	}
	return ret
}

func (c *Context) PageSize() int {
	page := c.Context.DefaultQuery("size", "20")
	ret, err := strconv.Atoi(page)
	if err != nil {
		return 20
	}
	return ret
}

func (c *Context) ShouldBindJSON(out any) error {
	err := c.Context.ShouldBindJSON(out)
	if err != nil {
		return err
	}
	return validate.Struct(out)
}

func (c *Context) WriteFail(code int, message string) {
	c.AbortWithStatusJSON(200, &Response{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().Local().Unix(),
	})
}

func (c *Context) WriteResponse(res *Response) {
	res.Timestamp = time.Now().Local().Unix()
	c.AbortWithStatusJSON(200, res)
}

func (c *Context) Write(code int, data any) {
	c.AbortWithStatusJSON(200, data)
}

func (c *Context) WriteData(data any) {
	c.Write(200, &Response{
		Code:      0,
		Content:   data,
		Timestamp: time.Now().Local().Unix(),
	})
}

func (c *Context) WriteDataWithPagination(data any, pagination any) {
	c.AbortWithStatusJSON(200, &Response{
		Code:       0,
		Content:    data,
		Pagination: pagination,
		Timestamp:  time.Now().Local().Unix(),
	})
}
