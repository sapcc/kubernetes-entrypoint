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
	"strings"

	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	discoveryv1applyconfigurations "k8s.io/client-go/applyconfigurations/discovery/v1"
	discoveryv1client "k8s.io/client-go/kubernetes/typed/discovery/v1"
)

type esClient struct{}

const (
	MockEndpointError = "mock endpoint didnt work"
)

func boolPtr(b bool) *bool { return &b }

func (e esClient) Create(
	ctx context.Context,
	endpointSlice *discoveryv1.EndpointSlice,
	opts metav1.CreateOptions,
) (*discoveryv1.EndpointSlice, error) {

	return nil, errors.New("not implemented")
}

func (e esClient) Update(
	ctx context.Context,
	endpointSlice *discoveryv1.EndpointSlice,
	opts metav1.UpdateOptions,
) (*discoveryv1.EndpointSlice, error) {

	return nil, errors.New("not implemented")
}

func (e esClient) Delete(
	ctx context.Context,
	name string,
	opts metav1.DeleteOptions,
) error {

	return errors.New("not implemented")
}

func (e esClient) DeleteCollection(
	ctx context.Context,
	opts metav1.DeleteOptions,
	listOpts metav1.ListOptions,
) error {

	return errors.New("not implemented")
}

func (e esClient) Get(
	ctx context.Context,
	name string,
	opts metav1.GetOptions,
) (*discoveryv1.EndpointSlice, error) {

	return nil, errors.New("not implemented")
}

func (e esClient) List(
	ctx context.Context,
	opts metav1.ListOptions,
) (*discoveryv1.EndpointSliceList, error) {

	// Extract service name from label selector "kubernetes.io/service-name=<name>"
	name := strings.TrimPrefix(opts.LabelSelector, "kubernetes.io/service-name=")

	if name == FailingServiceName {
		return nil, errors.New(MockEndpointError)
	}

	var endpoints []discoveryv1.Endpoint

	if name != EmptySubsetsServiceName {
		ready := boolPtr(true)
		endpoints = []discoveryv1.Endpoint{
			{
				Addresses:  []string{"127.0.0.1"},
				Conditions: discoveryv1.EndpointConditions{Ready: ready},
			},
		}
	}

	sliceList := &discoveryv1.EndpointSliceList{
		Items: []discoveryv1.EndpointSlice{
			{
				ObjectMeta: metav1.ObjectMeta{Name: name},
				Endpoints:  endpoints,
			},
		},
	}

	return sliceList, nil
}

func (e esClient) Watch(
	ctx context.Context,
	opts metav1.ListOptions,
) (watch.Interface, error) {

	return nil, errors.New("not implemented")
}

func (e esClient) Patch(
	ctx context.Context,
	name string,
	pt types.PatchType,
	data []byte,
	opts metav1.PatchOptions,
	subresources ...string,
) (result *discoveryv1.EndpointSlice, err error) {

	return nil, errors.New("not implemented")
}

func (e esClient) Apply(
	ctx context.Context,
	endpointSlice *discoveryv1applyconfigurations.EndpointSliceApplyConfiguration,
	opts metav1.ApplyOptions,
) (result *discoveryv1.EndpointSlice, err error) {

	return nil, errors.New("not implemented")
}

func NewEClient() discoveryv1client.EndpointSliceInterface {
	return esClient{}
}