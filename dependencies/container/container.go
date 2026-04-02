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

package container

import (
	"context"
	"errors"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	entry "opendev.org/airship/kubernetes-entrypoint/entrypoint"
	"opendev.org/airship/kubernetes-entrypoint/logger"
	"opendev.org/airship/kubernetes-entrypoint/util"
	"opendev.org/airship/kubernetes-entrypoint/util/env"
)

const (
	PodNameNotSetError    = "environment variable POD_NAME not set"
	NamespaceNotSupported = "container doesn't accept namespace"
)

type Container struct {
	name string
}

func init() {
	containerEnv := entry.DependencyPrefix + "CONTAINER"
	if util.ContainsSeparator(containerEnv, "Container") {
		logger.Error.Print(NamespaceNotSupported)
		os.Exit(1)
	}
	if containerDeps := env.SplitEnvToDeps(containerEnv); containerDeps != nil {
		if len(containerDeps) > 0 {
			for _, dep := range containerDeps {
				entry.Register(NewContainer(dep.Name))
			}
		}
	}
}

func NewContainer(name string) Container {
	return Container{name: name}
}

func (c Container) IsResolved(ctx context.Context, entrypoint entry.EntrypointInterface) (bool, error) {
	myPodName := os.Getenv("POD_NAME")
	if myPodName == "" {
		return false, errors.New(PodNameNotSetError)
	}
	pod, err := entrypoint.Client().Pods(env.GetBaseNamespace()).Get(ctx, myPodName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	if strings.Contains(c.name, env.Separator) {
		return false, errors.New("specifying namespace is not permitted")
	}
	containers := pod.Status.ContainerStatuses
	for _, container := range containers {
		if container.Name == c.name && container.Ready {
			return true, nil
		}
	}
	return false, nil
}

func (c Container) String() string {
	return "Container " + c.name
}
