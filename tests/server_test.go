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

package tests

import (
	"context"
	"fmt"
	"github.com/xfali/neve-websocket/server"
	websocket2 "github.com/xfali/neve-websocket/websocket"
	"net/http"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	handler := server.NewHandler()
	handler.RegisterConnectListener(&serverConnListener{t: t})
	http.ListenAndServe(":8080", handler)
}

type serverConnListener struct {
	t *testing.T
}

func (l *serverConnListener) OnNewConnect(ctx context.Context, r *http.Request, channel websocket2.MessageChannel) {
	l.t.Log("Client Connected ", r.RemoteAddr)
	reader := channel.GetReadChan()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case v := <-reader:
				l.t.Log(string(v))
			}
		}
	}()

	writer := channel.GetWriteChan()
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case v := <-reader:
				l.t.Log(string(v))
			case <-ticker.C:
				writer <- []byte(fmt.Sprintf("server: %v", time.Now()))
			}
		}
	}()
}
