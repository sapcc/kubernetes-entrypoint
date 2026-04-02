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

package mocks

import (
	"context"
	"errors"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	appsv1applyconfigurations "k8s.io/client-go/applyconfigurations/apps/v1"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

type dClient struct{}

const (
	SucceedingDaemonsetName         = "DAEMONSET_SUCCEED"
	FailingDaemonsetName            = "DAEMONSET_FAIL"
	CorrectNamespaceDaemonsetName   = "CORRECT_DAEMONSET_NAMESPACE"
	IncorrectNamespaceDaemonsetName = "INCORRECT_DAEMONSET_NAMESPACE"
	CorrectDaemonsetNamespace       = "CORRECT_DAEMONSET"

	FailingMatchLabelsDaemonsetName  = "DAEMONSET_INCORRECT_MATCH_LABELS"
	NotReadyMatchLabelsDaemonsetName = "DAEMONSET_NOT_READY_MATCH_LABELS"
)

func (d dClient) Create(
	ctx context.Context,
	daemonSet *v1.DaemonSet,
	opts metav1.CreateOptions,
) (*v1.DaemonSet, error) {

	return nil, errors.New("not implemented")
}

func (d dClient) Update(
	ctx context.Context,
	daemonSet *v1.DaemonSet,
	opts metav1.UpdateOptions,
) (*v1.DaemonSet, error) {

	return nil, errors.New("not implemented")
}

func (d dClient) UpdateStatus(
	ctx context.Context,
	daemonSet *v1.DaemonSet,
	opts metav1.UpdateOptions,
) (*v1.DaemonSet, error) {

	return nil, errors.New("not implemented")
}

func (d dClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return errors.New("not implemented")
}

func (d dClient) DeleteCollection(
	ctx context.Context,
	opts metav1.DeleteOptions,
	listOpts metav1.ListOptions,
) error {

	return errors.New("not implemented")
}

func (d dClient) Get(
	ctx context.Context,
	name string,
	opts metav1.GetOptions,
) (*v1.DaemonSet, error) {

	if name == FailingDaemonsetName || name == IncorrectNamespaceDaemonsetName {
		return nil, errors.New("mock daemonset didn't work")
	}

	matchLabelName := MockContainerName
	switch name {
	case FailingMatchLabelsDaemonsetName:
		matchLabelName = FailingMatchLabel
	case NotReadyMatchLabelsDaemonsetName:
		matchLabelName = SameHostNotReadyMatchLabel
	}

	ds := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"name": matchLabelName},
			},
		},
	}

	if name == CorrectNamespaceDaemonsetName {
		ds.Namespace = CorrectDaemonsetNamespace
	}

	return ds, nil
}

func (d dClient) List(ctx context.Context, opts metav1.ListOptions) (*v1.DaemonSetList, error) {
	return nil, errors.New("not implemented")
}

func (d dClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, errors.New("not implemented")
}

func (d dClient) Patch(
	ctx context.Context,
	name string,
	pt types.PatchType,
	data []byte,
	opts metav1.PatchOptions,
	subresources ...string,
) (result *v1.DaemonSet, err error) {

	return nil, errors.New("not implemented")
}

func (d dClient) Apply(
	ctx context.Context,
	daemonSet *appsv1applyconfigurations.DaemonSetApplyConfiguration,
	opts metav1.ApplyOptions,
) (result *v1.DaemonSet, err error) {

	return nil, errors.New("not implemented")
}

func (d dClient) ApplyStatus(
	ctx context.Context,
	daemonSet *appsv1applyconfigurations.DaemonSetApplyConfiguration,
	opts metav1.ApplyOptions,
) (result *v1.DaemonSet, err error) {

	return nil, errors.New("not implemented")
}

func NewDSClient() appsv1.DaemonSetInterface {
	return dClient{}
}
