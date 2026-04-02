// Copyright 2018 The kubernetes-entrypoint Authors
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

package entrypoint

import (
	"context"
	"sync"
	"time"

	"k8s.io/client-go/rest"

	"opendev.org/airship/kubernetes-entrypoint/client"
	"opendev.org/airship/kubernetes-entrypoint/logger"
)

var dependencies []Resolver // List containing all dependencies to be resolved
const (
	// DependencyPrefix is a prefix for env variables
	DependencyPrefix      = "DEPENDENCY_"
	JsonSuffix            = "_JSON"
	resolverSleepInterval = 2
)

// Resolver is an interface which all dependencies should implement
type Resolver interface {
	IsResolved(ctx context.Context, entrypoint EntrypointInterface) (bool, error)
}

type EntrypointInterface interface {
	Resolve()
	Client() client.ClientInterface
}

// Entrypoint is a main struct which checks dependencies
type Entrypoint struct {
	client    client.ClientInterface
	namespace string
}

// Register is a function which registers new dependencies
func Register(res Resolver) {
	if res == nil {
		panic("Entrypoint: could not register nil Resolver")
	}
	dependencies = append(dependencies, res)
}

// New is a constructor for entrypoint
func New(config *rest.Config) (entry *Entrypoint, err error) {
	entry = new(Entrypoint)
	newClient, err := client.New(config)
	if err != nil {
		return nil, err
	}
	entry.client = newClient
	return entry, err
}

func (e Entrypoint) Client() (c client.ClientInterface) {
	return e.client
}

// Resolve is a main loop which iterates through all dependencies and resolves them
func (e Entrypoint) Resolve() {
	ctx := context.Background()
	var wg sync.WaitGroup
	for _, dep := range dependencies {
		wg.Add(1)
		go func(dep Resolver) {
			defer wg.Done()
			logger.Info.Printf("Resolving %v", dep)
			var err error
			status := false
			for !status {
				if status, err = dep.IsResolved(ctx, e); err != nil {
					logger.Warning.Printf("Resolving dependency %+v failed: %v .", dep, err)
				}
				time.Sleep(resolverSleepInterval * time.Second)
			}
			logger.Info.Printf("Dependency %v is resolved.", dep)
		}(dep)
	}
	wg.Wait()
}
