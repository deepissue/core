package server

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Name  string
	Tags  []string
	Func  func(*Context) error
	Args  any
	Reply any
}

type APIHandler interface {
	Get(path string, handler *Handler)
	Post(path string, handler *Handler)
	Put(path string, handler *Handler)
	Delete(path string, handler *Handler)
	Head(path string, handler *Handler)
	Options(path string, handler *Handler)

	Handle(method string, path string, handler *Handler)

	Internal(method string, path string, handler *Handler)
}

// Handle registers a new route with the HTTP server.
func (m *HttpServer) Handle(method string, path string, handler *Handler) {

	path, _ = url.JoinPath(m.path, path)

	m.engine.Handle(method, path, func(c *gin.Context) {
		ctx := NewContext(c)
		// if err := m.Authorization(ctx); err != nil {
		// 	ctx.Status(401)
		// 	ctx.Writer.WriteString(err.Error())
		// 	return
		// }
		err := handler.Func(ctx)
		if nil != err {
			ctx.Status(400)
			ctx.Writer.WriteString(err.Error())
		}
	})
	path = strings.TrimPrefix(path, "/")
	m.addHandlerDoc(method, "/"+path, handler)
}

func (m *HttpServer) Internal(method string, path string, handler *Handler) {
	path, _ = url.JoinPath(m.path, "internal", path)

	m.engine.Handle(method, path, func(c *gin.Context) {
		ctx := NewContext(c)
		if c.GetHeader(InternalSecretKey) != m.authorization.Settings().InternalSecret {
			ctx.WriteFail(401, "Internal secret key required")
			return
		}
		err := handler.Func(ctx)
		if nil != err {
			ctx.Status(400)
			ctx.Writer.WriteString(err.Error())
		}
	})
	path = strings.TrimPrefix(path, "/")
	m.addHandlerDoc(method, "/"+path, handler)
}

func (m *HttpServer) Get(path string, handler *Handler) {
	m.Handle(http.MethodGet, path, handler)
}

func (m *HttpServer) Post(path string, handler *Handler) {
	m.Handle(http.MethodPost, path, handler)
}

func (m *HttpServer) Put(path string, handler *Handler) {
	m.Handle(http.MethodPut, path, handler)
}
func (m *HttpServer) Delete(path string, handler *Handler) {
	m.Handle(http.MethodDelete, path, handler)
}
func (m *HttpServer) Head(path string, handler *Handler) {
	m.Handle(http.MethodHead, path, handler)
}
func (m *HttpServer) Options(path string, handler *Handler) {
	m.Handle(http.MethodOptions, path, handler)
}
func (m *HttpServer) Patch(path string, handler *Handler) {
	m.Handle(http.MethodPatch, path, handler)
}
