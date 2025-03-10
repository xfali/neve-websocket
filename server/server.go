/*
 * Copyright (C) 2024, Xiongfa Li.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"context"
	"github.com/gorilla/websocket"
	websocket2 "github.com/xfali/neve-websocket/websocket"
	"github.com/xfali/xlog"
	"net/http"
	"sync"
)

type ResponseHeaderReader func(r *http.Request) http.Header
type ErrorWriter func(w http.ResponseWriter, err error)
type ConnectListener interface {
	OnNewConnect(ctx context.Context, r *http.Request, channel websocket2.MessageChannel)
}

type Opt func(*Handler)

type Handler struct {
	logger   xlog.Logger
	stopCtx  context.Context
	stopFunc context.CancelFunc

	upgrader *websocket.Upgrader

	responseHeaderReader ResponseHeaderReader
	errorProcessor       ErrorWriter

	channelFac websocket2.MessageChannelFactory

	connListeners    []ConnectListener
	connListenerLock sync.RWMutex
}

func DefaultErrorWriter(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}

func NewHandler(opts ...Opt) *Handler {
	ret := &Handler{
		logger: xlog.GetLogger(),
		responseHeaderReader: func(r *http.Request) http.Header {
			return nil
		},
	}
	ret.errorProcessor = ret.errorWriter
	ret.channelFac = ret
	ret.upgrader = &websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	ret.stopCtx, ret.stopFunc = context.WithCancel(context.Background())

	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (o *Handler) Close() error {
	o.stopFunc()
	return nil
}

func (o *Handler) BeanDestroy() error {
	return o.Close()
}

func (o *Handler) CreateMessageChannel(ctx context.Context) (websocket2.MessageChannel, error) {
	return websocket2.NewMessageChannel(make(chan []byte, 4096), make(chan []byte, 4096)), nil
}

func (o *Handler) RegisterConnectListener(listener ...ConnectListener) {
	o.connListenerLock.Lock()
	defer o.connListenerLock.Unlock()

	o.connListeners = append(o.connListeners, listener...)
}

func (o *Handler) notifyConnect(r *http.Request, ch websocket2.MessageChannel) {
	o.connListenerLock.RLock()
	defer o.connListenerLock.RUnlock()

	for _, l := range o.connListeners {
		l.OnNewConnect(o.stopCtx, r, ch)
	}
}

func (o *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	o.Ws(w, r)
}

func (o *Handler) Ws(w http.ResponseWriter, r *http.Request) {
	conn, err := o.upgrader.Upgrade(w, r, o.responseHeaderReader(r))
	if err != nil {
		o.errorProcessor(w, err)
		return
	}

	defer conn.Close()

	ch, err := o.channelFac.CreateMessageChannel(r.Context())
	if err != nil {
		o.errorProcessor(w, err)
		return
	}

	o.notifyConnect(r, ch)
	err = ch.Listen(o.stopCtx, conn)
	if err != nil {
		o.errorProcessor(w, err)
		return
	}
}

func (o *Handler) errorWriter(w http.ResponseWriter, err error) {
	o.logger.Errorln(err)
	http.Error(w, err.Error(), http.StatusBadRequest)
}

type opts struct {
}

var Opts opts

func (o opts) SetUpgrader(upgrader *websocket.Upgrader) Opt {
	return func(server *Handler) {
		server.upgrader = upgrader
	}
}

func (o opts) SetResponseHeaderReader(responseHeaderReader ResponseHeaderReader) Opt {
	return func(server *Handler) {
		server.responseHeaderReader = responseHeaderReader
	}
}

func (o opts) SetErrorWriter(errorProcessor ErrorWriter) Opt {
	return func(server *Handler) {
		server.errorProcessor = errorProcessor
	}
}

func (o opts) SetMessageChannelFactory(channelFac websocket2.MessageChannelFactory) Opt {
	return func(server *Handler) {
		server.channelFac = channelFac
	}
}

func (o opts) AddConnectListener(listeners ...ConnectListener) Opt {
	return func(server *Handler) {
		server.connListeners = append(server.connListeners, listeners...)
	}
}
