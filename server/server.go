/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package server

import (
	"context"
	"time"

	"github.com/deepissue/core/logging"
	"github.com/deepissue/core/option"
	"github.com/deepissue/core/utils"
	_ "github.com/go-sql-driver/mysql"
)

type Server struct {
	Ctx    context.Context
	opts   *option.Options
	cancel context.CancelFunc
	logger *logging.Logger
	doneCh chan struct{}
}

func NewServer(opts *option.Options, logger *logging.Logger) (*Server, error) {

	ctx, cancel := context.WithCancel(context.Background())
	srv := &Server{
		Ctx:    ctx,
		opts:   opts,
		cancel: cancel,
		logger: logger,
		doneCh: utils.MakeShutdownCh(),
	}
	return srv, nil
}

func (m *Server) HandleSignal(onStop func()) {
	<-m.doneCh
	onStop()
	m.cancel()
	time.Sleep(time.Second)
	m.logger.Info("server shutting...")
	m.logger.Info("server shutdown completed")
	m.logger.Cleanup()
}
