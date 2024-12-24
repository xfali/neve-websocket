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

package connection

import (
	"context"
	"github.com/gorilla/websocket"
	"time"
)

type Connection struct {
	conn      *websocket.Conn
	stopCtx   context.Context
	stopFunc  context.CancelFunc
	readChan  chan []byte
	writeChan chan []byte
}

func NewConnection(ctx context.Context, conn *websocket.Conn, readChan, writeChan chan []byte) *Connection {
	ret := &Connection{
		conn:      conn,
		readChan:  readChan,
		writeChan: writeChan,
	}
	ret.stopCtx, ret.stopFunc = context.WithCancel(ctx)
	return ret
}

func (c *Connection) Close() error {
	c.stopFunc()
	return nil
}

func (c *Connection) ReadLoop() error {
	if c.readChan == nil {
		return nil
	}
	exitCh := make(chan struct{})
	defer close(exitCh)

	go func() {
		select {
		case <-exitCh:
			return
		case <-c.stopCtx.Done():
			_ = c.conn.SetReadDeadline(time.Now())
		}
	}()

	for {
		select {
		case <-c.stopCtx.Done():
			return c.stopCtx.Err()
		default:
			_, p, err := c.conn.ReadMessage()
			if err != nil {
				return err
			}
			select {
			case <-c.stopCtx.Done():
				return c.stopCtx.Err()
			case c.readChan <- p:
			}
		}
	}
}

func (c *Connection) WriteLoop() error {
	if c.writeChan == nil {
		return nil
	}
	exitCh := make(chan struct{})
	defer close(exitCh)

	go func() {
		select {
		case <-exitCh:
			return
		case <-c.stopCtx.Done():
			_ = c.conn.SetWriteDeadline(time.Now())
		}
	}()

	for {
		select {
		case <-c.stopCtx.Done():
			return c.stopCtx.Err()
		case v := <-c.writeChan:
			err := c.conn.WriteMessage(websocket.TextMessage, v)
			if err != nil {
				return err
			}
		}
	}
}
