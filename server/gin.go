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
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/xfali/neve-core/appcontext"
	"github.com/xfali/neve-websocket/events"
	websocket2 "github.com/xfali/neve-websocket/websocket"
	"github.com/xfali/xlog"
	"reflect"
)

type websocketHandler struct {
	logger  xlog.Logger
	handler *Handler

	Path          string            `fig:"neve.web.websocket.path"`
	ConnListeners []ConnectListener `inject:""`
}

func NewWebsocketHandler(opts ...Opt) *websocketHandler {
	ret := &websocketHandler{
		logger:  xlog.GetLogger(),
		handler: NewHandler(opts...),
	}

	return ret
}

func (o *websocketHandler) BeanAfterSet() error {
	o.handler.RegisterConnectListener(o.ConnListeners...)
	return nil
}

func (o *websocketHandler) HttpRoutes(engine gin.IRouter) {
	if o.Path == "" {
		o.Path = "/ws"
	}
	engine.GET(o.Path, func(context *gin.Context) {
		o.handler.Ws(context.Writer, context.Request)
	})
}

type websocketReadEvent struct {
	WebsocketReadEventType string `json:"websocket_read_event_type"`
}

type websocketEventHandler struct {
	logger   xlog.Logger
	stopCtx  context.Context
	stopFunc context.CancelFunc

	handler *Handler

	channel   websocket2.MessageChannel
	readChan  chan []byte
	writeChan chan []byte

	Path      string                               `fig:"neve.web.websocket.event.path"`
	Publisher appcontext.ApplicationEventPublisher `inject:""`
}

func NewWebsocketEventHandler(readChanSize, writeChanSize int, opts ...Opt) *websocketEventHandler {
	ret := &websocketEventHandler{
		logger:  xlog.GetLogger(),
		handler: NewHandler(opts...),
	}
	ret.handler.channelFac = ret
	enableRead := false

	if readChanSize >= 0 {
		ret.readChan = make(chan []byte, readChanSize)
		enableRead = true
	}
	if writeChanSize >= 0 {
		ret.writeChan = make(chan []byte, writeChanSize)
	}

	ret.channel = websocket2.NewMessageChannel(ret.readChan, ret.writeChan)
	ret.stopCtx, ret.stopFunc = context.WithCancel(context.Background())
	if enableRead {
		go ret.readEventLoop()
	}

	return ret
}

func (o *websocketEventHandler) CreateMessageChannel(ctx context.Context) (websocket2.MessageChannel, error) {
	return o.channel, nil
}

func (o *websocketEventHandler) BeanDestroy() error {
	o.stopFunc()
	return nil
}

func (o *websocketEventHandler) HttpRoutes(engine gin.IRouter) {
	if o.Path == "" {
		o.Path = "/ws/events"
	}
	engine.GET(o.Path, func(context *gin.Context) {
		o.handler.Ws(context.Writer, context.Request)
	})
}

func (o *websocketEventHandler) readEventLoop() {
	for {
		select {
		case <-o.stopCtx.Done():
			return
		case v := <-o.readChan:
			e := &websocketReadEvent{}
			err := json.Unmarshal(v, e)
			if err == nil {
				if t, ok := events.GetType(e.WebsocketReadEventType); ok {
					target := reflect.New(t).Interface()
					if um, ok := target.(events.WebsocketReadEvent); ok {
						err = um.UnmarshalWebsocketEvent(v)
						if err == nil {
							_ = o.Publisher.SendEvent(um)
						}
					}
				} else {
					// Put back again?
					//o.readChan <- v
				}
			}
		}
	}
}

func (o *websocketEventHandler) RegisterConsumer(registry appcontext.ApplicationEventConsumerRegistry) error {
	return registry.RegisterApplicationEventConsumer(o.handlerEvent)
}

func (o *websocketEventHandler) handlerEvent(event events.WebsocketWriteEvent) {
	v, err := event.MarshalWebsocketEvent()
	if err != nil {
		o.logger.Errorln(err)
		return
	}
	select {
	case o.writeChan <- v:
		o.logger.Debugln("Write websocket event success")
	}
}
