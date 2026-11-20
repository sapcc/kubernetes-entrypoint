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

package deployment

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	entry "opendev.org/airship/kubernetes-entrypoint/entrypoint"
	"opendev.org/airship/kubernetes-entrypoint/logger"
	"opendev.org/airship/kubernetes-entrypoint/util/env"
)

type Deployment struct {
	name        string
	namespace   string
	annotations map[string]string
	labels      map[string]string
}

func init() {
	registerDeployments(entry.Register)
}

func registerDeployments(register func(entry.Resolver)) {
	deploymentEnv := entry.DependencyPrefix + "DEPLOYMENT"
	for _, dep := range env.SplitDeploymentEnvToDeps(deploymentEnv, deploymentEnv+entry.JsonSuffix) {
		if deployment := NewDeployment(dep.Name, dep.Namespace, dep.Annotations, dep.Labels); deployment != nil {
			register(*deployment)
		}
	}
}

func NewDeployment(name, namespace string, annotations, labels map[string]string) *Deployment {
	if strings.TrimSpace(name) == "" {
		logger.Warning.Printf("Cannot create Deployment dependency: name is required")
		return nil
	}
	return &Deployment{name: name, namespace: namespace, annotations: maps.Clone(annotations), labels: maps.Clone(labels)}
}

func (d Deployment) String() string {
	return fmt.Sprintf("Deployment %s in namespace %s", d.name, d.namespace)
}

func (d Deployment) IsResolved(ctx context.Context, entrypoint entry.EntrypointInterface) (bool, error) {
	snapshot, err := entrypoint.Client().Deployments(d.namespace).Get(ctx, d.name, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("getting %s: %w", d, err)
	}
	resolved, err := evaluateDeployment(snapshot, d.annotations, d.labels)
	if err != nil {
		return false, fmt.Errorf("%s: %w", d, err)
	}
	return resolved, nil
}

// evaluateDeployment uses only this snapshot: stale success cannot acknowledge
// a new specification. Completion requires replica convergence and no terminating
// replicas, rather than relying on conditions.
func evaluateDeployment(deployment *appsv1.Deployment, annotations, labels map[string]string) (bool, error) {
	if deployment == nil {
		return false, errors.New("Deployment was not returned by the API")
	}
	if deployment.DeletionTimestamp != nil {
		return false, errors.New("Deployment is being deleted")
	}
	if err := matchMetadata("annotation", deployment.Annotations, annotations); err != nil {
		return false, err
	}
	if err := matchMetadata("label", deployment.Labels, labels); err != nil {
		return false, err
	}
	status := deployment.Status
	if status.ObservedGeneration < deployment.Generation {
		return false, fmt.Errorf("generation %d has not been observed (observed generation %d)", deployment.Generation, status.ObservedGeneration)
	}
	if deployment.Spec.Paused {
		return false, errors.New("Deployment is paused")
	}
	desired := int32(1)
	if deployment.Spec.Replicas != nil {
		desired = *deployment.Spec.Replicas
	}
	if status.TerminatingReplicas != nil && *status.TerminatingReplicas != 0 {
		return false, errors.New("rollout incomplete: terminating Pods remain")
	}
	if status.UpdatedReplicas == desired && status.Replicas == desired && status.AvailableReplicas == desired {
		return true, nil
	}

	return false, errors.New("rollout incomplete")
}

func matchMetadata(kind string, actual, expected map[string]string) error {
	for key, value := range expected {
		got, present := actual[key]
		if !present {
			return fmt.Errorf("pending target update: metadata %s %q is missing, expected %q", kind, key, value)
		}
		if got != value {
			return fmt.Errorf("pending target update: metadata %s %q is %q, expected %q", kind, key, got, value)
		}
	}
	return nil
}
