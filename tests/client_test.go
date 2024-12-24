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
	"github.com/xfali/neve-websocket/client"
	websocket2 "github.com/xfali/neve-websocket/websocket"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	cli := client.NewClient("ws://localhost:8080/ws")
	cli.RegisterConnectListener(newClientConnListener(t))
	err := cli.Connect(context.Background())
	if err != nil {
		t.Log(err)
	}
}

func TestClient2(t *testing.T) {
	cli := client.NewClient("ws://localhost:8080/ws")
	cli.RegisterConnectListener(newClientConnListener(t))
	err := cli.Connect(context.Background())
	if err != nil {
		t.Log(err)
	}
}

type clientConnListener struct {
	t  *testing.T
	id int64
}

func newClientConnListener(t *testing.T) *clientConnListener {
	ret := &clientConnListener{t: t}
	ret.id = time.Now().UnixNano()
	return ret
}

func (l *clientConnListener) OnNewConnect(ctx context.Context, endpoint string, channel websocket2.MessageChannel) {
	l.t.Log("Connected ", endpoint)
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
				writer <- []byte(fmt.Sprintf("client %d : %v", l.id, time.Now()))
			}
		}
	}()
}
