// Copyright 2017 The kubernetes-entrypoint Authors
// Copyright 2026 SAP SE or an SAP affiliate company
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

package socket

import (
	"context"
	"fmt"
	"os"

	entry "opendev.org/airship/kubernetes-entrypoint/entrypoint"
	"opendev.org/airship/kubernetes-entrypoint/logger"
	"opendev.org/airship/kubernetes-entrypoint/util"
	"opendev.org/airship/kubernetes-entrypoint/util/env"
)

const (
	NonExistingErrorFormat = "%s doesn't exists"
	NoPermsErrorFormat     = "no permission to %s"
	NamespaceNotSupported  = "socket doesn't accept namespace"
)

type Socket struct {
	name string
}

func init() {
	socketEnv := entry.DependencyPrefix + "SOCKET"
	if util.ContainsSeparator(socketEnv, "Socket") {
		logger.Error.Print(NamespaceNotSupported)
		os.Exit(1)
	}
	if socketDeps := env.SplitEnvToDeps(socketEnv); socketDeps != nil {
		if len(socketDeps) > 0 {
			for _, dep := range socketDeps {
				entry.Register(NewSocket(dep.Name))
			}
		}
	}
}

func NewSocket(name string) Socket {
	return Socket{name: name}
}

func (s Socket) IsResolved(ctx context.Context, entrypoint entry.EntrypointInterface) (bool, error) {
	_, err := os.Stat(s.name)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, fmt.Errorf(NonExistingErrorFormat, s)
	}
	if os.IsPermission(err) {
		return false, fmt.Errorf(NoPermsErrorFormat, s)
	}
	return false, err
}

func (s Socket) String() string {
	return "Socket " + s.name
}
