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

package websocket

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/xfali/neve-websocket/connection"
	"github.com/xfali/xlog"
	"sync"
)

type defaultChannel struct {
	logger    xlog.Logger
	readChan  chan []byte
	writeChan chan []byte
	conn      *connection.Connection
}

func NewMessageChannel(readChan chan []byte, writeChan chan []byte) *defaultChannel {
	return &defaultChannel{
		logger:    xlog.GetLogger(),
		readChan:  readChan,
		writeChan: writeChan,
	}
}

func (o *defaultChannel) Close() error {
	if o.conn != nil {
		err := o.conn.Close()
		if err != nil {
			return err
		}
		o.conn = nil
	}
	return nil
}

func (o *defaultChannel) Listen(ctx context.Context, conn *websocket.Conn) error {
	if o.conn != nil {
		return fmt.Errorf("Already Listening ")
	}
	connect := connection.NewConnection(ctx, conn, o.readChan, o.writeChan)
	defer connect.Close()
	o.conn = connect

	wait := sync.WaitGroup{}
	if o.writeChan != nil {
		wait.Add(1)
		go func() {
			defer wait.Done()
			err := connect.WriteLoop()
			if err != nil {
				o.logger.Errorln("Write error: ", err)
			}
		}()
	}

	if o.readChan != nil {
		wait.Add(1)
		go func() {
			defer wait.Done()
			err := connect.ReadLoop()
			if err != nil {
				o.logger.Errorln("Read error: ", err)
			}
		}()
	}

	wait.Wait()
	return nil
}

func (o *defaultChannel) GetReadChan() <-chan []byte {
	return o.readChan
}

func (o *defaultChannel) GetWriteChan() chan<- []byte {
	return o.writeChan
}
