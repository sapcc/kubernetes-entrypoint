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

package job

import (
	"context"
	"fmt"

	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	entry "opendev.org/airship/kubernetes-entrypoint/entrypoint"
	"opendev.org/airship/kubernetes-entrypoint/logger"
	"opendev.org/airship/kubernetes-entrypoint/util/env"
)

const FailingStatusFormat = "job %s is not completed yet"

type Job struct {
	name      string
	namespace string
	labels    map[string]string
}

func init() {
	jobsEnv := entry.DependencyPrefix + "JOBS"
	jobsJsonEnv := fmt.Sprintf("%s%s", jobsEnv, entry.JsonSuffix)
	if jobsDeps := env.SplitJobEnvToDeps(jobsEnv, jobsJsonEnv); jobsDeps != nil {
		if len(jobsDeps) > 0 {
			for _, dep := range jobsDeps {
				job := NewJob(dep.Name, dep.Namespace, dep.Labels)
				if job != nil {
					entry.Register(*job)
				}
			}
		}
	}
}

func NewJob(name, namespace string, labels map[string]string) *Job {
	if name != "" && labels != nil {
		logger.Warning.Printf("Cannot specify both name and labels for job depependency")
		return nil
	}
	return &Job{
		name:      name,
		namespace: namespace,
		labels:    labels,
	}
}

func (j Job) IsResolved(ctx context.Context, entrypoint entry.EntrypointInterface) (bool, error) {
	iface := entrypoint.Client().Jobs(j.namespace)
	var jobs []v1.Job

	if j.name != "" {
		job, err := iface.Get(ctx, j.name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		jobs = []v1.Job{*job}
	} else if j.labels != nil {
		labelSelector := &metav1.LabelSelector{MatchLabels: j.labels}
		label := metav1.FormatLabelSelector(labelSelector)
		opts := metav1.ListOptions{LabelSelector: label}
		jobList, err := iface.List(ctx, opts)
		if err != nil {
			return false, err
		}
		jobs = jobList.Items
	}
	if len(jobs) == 0 {
		return false, fmt.Errorf("no matching jobs found: %v", j)
	}

	for _, job := range jobs {
		if job.Status.Succeeded == 0 {
			return false, fmt.Errorf(FailingStatusFormat, j)
		}
	}
	return true, nil
}

func (j Job) String() string {
	prefix := "Jobs"
	if j.name != "" {
		prefix = "Job " + j.name
	} else if j.labels != nil {
		prefix = fmt.Sprintf("Jobs with labels %s", j.labels)
	}
	return fmt.Sprintf("%s in namespace %s", prefix, j.namespace)
}
