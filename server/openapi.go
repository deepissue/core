package server

import (
	"log"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func (m *HttpServer) openapi(path string) {
	path, _ = url.JoinPath(m.path, path)
	m.engine.Handle("GET", path, func(ctx *gin.Context) {

		if strings.HasSuffix(path, "json") {
			schema, err := m.reflector.Spec.MarshalJSON()
			if err != nil {
				log.Fatal(err)
			}
			ctx.Header("content-type", "application/json")
			ctx.Writer.Write(schema)
		}

		if strings.HasSuffix(path, "yaml") || strings.HasSuffix(path, "yml") {
			schema, err := m.reflector.Spec.MarshalYAML()
			if err != nil {
				log.Fatal(err)
			}
			ctx.Header("content-type", "application/yaml")
			ctx.Writer.Write(schema)
		}

	})
}

func (m *HttpServer) addHandlerDoc(method string, path string, handler *Handler) {

	operation, _ := m.reflector.NewOperationContext(method, path)
	operation.SetSummary(handler.Name)
	operation.SetTags(handler.Tags...)
	operation.AddReqStructure(handler.Args)
	operation.AddRespStructure(handler.Reply)

	m.reflector.AddOperation(operation)

}
