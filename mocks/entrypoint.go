// Copyright 2017 The kubernetes-entrypoint Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mocks

import (
	"opendev.org/airship/kubernetes-entrypoint/client"
)

type MockEntrypoint struct {
	MockClient *Client
	namespace  string
}

func (m MockEntrypoint) Resolve() {}

func (m MockEntrypoint) Client() client.ClientInterface {
	return m.MockClient
}

func (m MockEntrypoint) GetNamespace() string {
	return m.namespace
}

func NewEntrypointInNamespace(namespace string) *MockEntrypoint {
	return &MockEntrypoint{
		MockClient: NewClient(),
		namespace:  namespace,
	}
}

func NewEntrypoint() *MockEntrypoint {
	return NewEntrypointInNamespace("test")
}
