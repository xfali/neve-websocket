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

package events

import (
	"reflect"
	"sync"
)

var (
	globalManager = NewManager()
)

type Manager struct {
	types      map[string]reflect.Type
	typeLocker sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		types: map[string]reflect.Type{},
	}
}

func (m *Manager) RegisterType(s string, t reflect.Type) {
	m.typeLocker.Lock()
	defer m.typeLocker.Unlock()

	m.types[s] = t
}

func (m *Manager) GetType(s string) (reflect.Type, bool) {
	m.typeLocker.RLock()
	defer m.typeLocker.RUnlock()

	v, ok := m.types[s]
	return v, ok
}

func RegisterType(s string, t reflect.Type) {
	globalManager.RegisterType(s, t)
}

func GetType(s string) (reflect.Type, bool) {
	return globalManager.GetType(s)
}
