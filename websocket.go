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

package nevewebsocket

import (
	"github.com/xfali/neve-websocket/client"
	"github.com/xfali/neve-websocket/server"
)

func NewServerHandler(opts ...server.Opt) *server.Handler {
	return server.NewHandler(opts...)
}

func NewWebsocketHandler(opts ...server.Opt) interface{} {
	return server.NewWebsocketHandler(opts...)
}

func NewWebsocketEventHandler(readChanSize, writeChanSize int, opts ...server.Opt) interface{} {
	return server.NewWebsocketEventHandler(readChanSize, writeChanSize, opts...)
}

func NewClient(endpoint string, opts ...client.Opt) *client.Client {
	return client.NewClient(endpoint, opts...)
}
