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

package client

import (
	"context"
	"github.com/gorilla/websocket"
	websocket2 "github.com/xfali/neve-websocket/websocket"
	"github.com/xfali/xlog"
	"net/http"
	"sync"
	"time"
)

type RequestHeaderReader func(endpoint string) http.Header
type ConnectListener interface {
	OnNewConnect(endpoint string, channel websocket2.MessageChannel)
}

type Opt func(*Client)

type Client struct {
	logger   xlog.Logger
	stopCtx  context.Context
	stopFunc context.CancelFunc

	endpoint      string
	retryInterval time.Duration

	dialer          *websocket.Dialer
	reqHeaderReader RequestHeaderReader

	channelFac websocket2.MessageChannelFactory

	connListeners    []ConnectListener
	connListenerLock sync.RWMutex
}

func NewClient(endpoint string, opts ...Opt) *Client {
	ret := &Client{
		endpoint:      endpoint,
		retryInterval: 10 * time.Second,
		dialer:        websocket.DefaultDialer,
		reqHeaderReader: func(endpoint string) http.Header {
			return nil
		},
		logger: xlog.GetLogger(),
	}
	ret.channelFac = ret
	ret.stopCtx, ret.stopFunc = context.WithCancel(context.Background())

	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (o *Client) Close() error {
	o.stopFunc()
	return nil
}

func (o *Client) BeanDestroy() error {
	return o.Close()
}

func (o *Client) CreateMessageChannel(ctx context.Context) (websocket2.MessageChannel, error) {
	return websocket2.NewMessageChannel(make(chan []byte, 4096), make(chan []byte, 4096)), nil
}

func (o *Client) RegisterConnectListener(listener ConnectListener) {
	o.connListenerLock.Lock()
	defer o.connListenerLock.Unlock()

	o.connListeners = append(o.connListeners, listener)
}

func (o *Client) notifyConnect(endpoint string, ch websocket2.MessageChannel) {
	o.connListenerLock.RLock()
	defer o.connListenerLock.RUnlock()

	for _, l := range o.connListeners {
		l.OnNewConnect(endpoint, ch)
	}
}

func (o *Client) Connect(ctx context.Context) error {
	o.stopCtx, o.stopFunc = context.WithCancel(ctx)
	err := o.connect(o.stopCtx)
	if err != nil {
		return err
	}
	timer := time.NewTimer(o.retryInterval)
	for {
		select {
		case <-o.stopCtx.Done():
			return o.stopCtx.Err()
		case <-timer.C:
			err = o.connect(o.stopCtx)
			if err != nil {
				o.logger.Errorln(err)
			}
			timer.Reset(o.retryInterval)
		}
	}
}

func (o *Client) connect(ctx context.Context) error {
	conn, _, err := o.dialer.DialContext(ctx, o.endpoint, o.reqHeaderReader(o.endpoint))
	if err != nil {
		o.logger.Errorln(err)
		return err
	}

	defer conn.Close()

	ch, err := o.channelFac.CreateMessageChannel(ctx)
	if err != nil {
		o.logger.Errorln(err)
		return err
	}

	o.notifyConnect(o.endpoint, ch)
	err = ch.Listen(o.stopCtx, conn)
	if err != nil {
		o.logger.Errorln(err)
	}
	return err
}

type opts struct {
}

var Opts opts

/*

 */
func (o opts) SetRetryInterval(retryInterval time.Duration) Opt {
	return func(client *Client) {
		client.retryInterval = retryInterval
	}
}

func (o opts) SetDialer(dialer *websocket.Dialer) Opt {
	return func(client *Client) {
		client.dialer = dialer
	}
}

func (o opts) SetRequestHeaderReader(reqHeaderReader RequestHeaderReader) Opt {
	return func(client *Client) {
		client.reqHeaderReader = reqHeaderReader
	}
}

func (o opts) SetMessageChannelFactory(channelFac websocket2.MessageChannelFactory) Opt {
	return func(client *Client) {
		client.channelFac = channelFac
	}
}

func (o opts) AddConnectListener(connListeners ...ConnectListener) Opt {
	return func(client *Client) {
		client.connListeners = append(client.connListeners, connListeners...)
	}
}
