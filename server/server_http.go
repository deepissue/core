/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/deepissue/core/authorities"
	"github.com/deepissue/core/option"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-hclog"
	"github.com/swaggest/openapi-go/openapi3"
)

type HttpServer struct {
	ctx           context.Context
	addr          string
	path          string
	ln            net.Listener
	logger        hclog.Logger
	engine        *gin.Engine
	httpServer    *http.Server
	authorization authorities.Authorization
	reflector     *openapi3.Reflector
}

func (m *Server) NewHttpServer(authorization authorities.Authorization) (*HttpServer, error) {
	if nil == authorization {
		return nil, errors.New("authorization can not be nil")
	}

	gin.DisableConsoleColor()
	recover := gin.RecoveryWithWriter(m.logger.StandardWriter(&hclog.StandardLoggerOptions{}))
	engine := gin.New()
	engine.Use(recover)
	engine.Use(HclogMiddleware(m.logger))
	addr := fmt.Sprintf("%s:%d", m.opts.Http.Address, m.opts.Http.Port)
	if m.opts.Http.Cors {
		engine.Use(Cors(m.opts))
	}

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ErrorLog:     m.logger.StandardLogger(&hclog.StandardLoggerOptions{}),
		ReadTimeout:  time.Duration(m.opts.Http.ReadTimeout) * time.Second,
		IdleTimeout:  time.Duration(m.opts.Http.IdleTimeout) * time.Second,
		WriteTimeout: time.Duration(m.opts.Http.WriteTimeout) * time.Second,
	}
	if m.opts.Http.Path != "" && m.opts.Http.Path[0] != '/' {
		return nil, errors.New("the http.path must start with a /")
	}

	reflector := openapi3.NewReflector()
	reflector.Spec = &openapi3.Spec{Openapi: "3.0.3"}
	reflector.Spec.Info.
		WithTitle(m.opts.Application + " API").WithDescription("")

	srv := &HttpServer{
		ctx:           m.Ctx,
		logger:        m.logger,
		engine:        engine,
		httpServer:    httpServer,
		addr:          addr,
		path:          m.opts.Http.Path,
		authorization: authorization,
		reflector:     reflector,
	}

	srv.openapi("openapi.json")
	return srv, nil
}

func (m *HttpServer) Engine() *gin.Engine {
	return m.engine
}

func (m *HttpServer) Use(middleware ...gin.HandlerFunc) *HttpServer {
	m.engine.Use(middleware...)
	return m
}

func (m *HttpServer) Startup() error {

	ln, err := net.Listen("tcp", m.addr)
	if err != nil {
		m.logger.Error("failed to listen on address", "addr", m.addr, "err", err)
		return err
	}
	m.logger.Info("http server listened on", "addr", m.addr)
	m.ln = ln
	go m.httpServer.Serve(m.ln)
	go func() {
		<-m.ctx.Done()
		m.Stop()
		m.ln.Close()
		m.httpServer.Close()
	}()
	return nil
}

func (m *HttpServer) Stop() {
	if err := m.httpServer.Shutdown(m.ctx); err != nil {
		m.logger.Error("shutdown http server", "err", "err")
		return
	}
}

func Cors(opts *option.Options) gin.HandlerFunc {
	headers := []string{
		"Origin", "Content-Kind",
		string(AuthorizationKey),
		string(InternalSecretKey), string(ClientIDKey),
		"Os-Version", "Application-Version", "Location", "Content-Disposition",
	}

	headers = append(headers, []string{
		"Sec-WebSocket-Extensions",
		"Sec-WebSocket-Version", "Sec-WebSocket-Key",
		"Sec-WebSocket-Protocol", "Upgrade",
	}...)

	return cors.New(cors.Config{
		//准许跨域请求网站,多个使用,分开,限制使用*
		AllowAllOrigins: true,
		//准许使用的请求方式
		AllowMethods: []string{"PUT", "PATCH", "POST", "GET", "DELETE", "OPTIONS"},
		//准许使用的请求表头
		AllowHeaders: []string{"*"},
		//显示的请求表头
		ExposeHeaders: headers,
		//凭证共享,确定共享
		AllowCredentials: true,
		//超时时间设定
		MaxAge:                 24 * time.Hour,
		AllowBrowserExtensions: true,
	})
}
