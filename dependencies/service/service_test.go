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

package service

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"opendev.org/airship/kubernetes-entrypoint/entrypoint"
	"opendev.org/airship/kubernetes-entrypoint/mocks"
)

const (
	testServiceName      = "TEST_SERVICE_NAME"
	testServiceNamespace = "TEST_SERVICE_NAMESPACE"
)

var testEntrypoint entrypoint.EntrypointInterface

var _ = Describe("Service", func() {
	BeforeEach(func() {
		testEntrypoint = mocks.NewEntrypoint()
	})

	It("checks the name of a newly created service", func() {
		service := NewService(testServiceName, testServiceNamespace)

		Expect(service.name).To(Equal(testServiceName))
		Expect(service.namespace).To(Equal(testServiceNamespace))
	})

	It("checks resolution of a succeeding service", func() {
		service := NewService(mocks.SucceedingServiceName, mocks.SucceedingServiceName)

		isResolved, err := service.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
	})

	It("checks resolution failure of a failing service", func() {
		service := NewService(mocks.FailingServiceName, mocks.FailingServiceName)

		isResolved, err := service.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeFalse())
		Expect(err.Error()).To(Equal(mocks.MockEndpointError))
	})

	It("checks resolution failure of a succeeding service with removed subsets", func() {
		service := NewService(mocks.EmptySubsetsServiceName, mocks.EmptySubsetsServiceName)

		isResolved, err := service.IsResolved(context.TODO(), testEntrypoint)
		Expect(isResolved).To(BeFalse())
		Expect(err.Error()).To(Equal(fmt.Sprintf(FailingStatusFormat, service.name)))
	})
})
