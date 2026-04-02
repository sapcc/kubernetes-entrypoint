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

package container

import (
	"context"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"opendev.org/airship/kubernetes-entrypoint/entrypoint"
	"opendev.org/airship/kubernetes-entrypoint/mocks"
)

const (
	podEnvVariableName = "POD_NAME"
)

var testEntrypoint entrypoint.EntrypointInterface

var _ = Describe("Container", func() {
	BeforeEach(func() {
		err := os.Setenv(podEnvVariableName, mocks.PodEnvVariableValue)
		Expect(err).NotTo(HaveOccurred())

		testEntrypoint = mocks.NewEntrypoint()
	})

	It("checks the name of a newly created container", func() {
		container := NewContainer(mocks.MockContainerName)

		Expect(container.name).To(Equal(mocks.MockContainerName))
	})

	It(fmt.Sprintf("checks container resolution failure with %s not set", podEnvVariableName), func() {
		err := os.Unsetenv(podEnvVariableName)
		Expect(err).NotTo(HaveOccurred())
		container := NewContainer(mocks.MockContainerName)

		isResolved, err := container.IsResolved(context.TODO(), testEntrypoint)
		Expect(isResolved).To(BeFalse())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(PodNameNotSetError))
	})

	It("checks resolution of a succeeding container", func() {
		container := NewContainer(mocks.MockContainerName)

		isResolved, err := container.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
	})

	It(fmt.Sprintf("fails to resolve a mocked container for a given %s value", podEnvVariableName), func() {
		err := os.Setenv(podEnvVariableName, "INVALID_POD_LIST_VALUE")
		Expect(err).NotTo(HaveOccurred())

		container := NewContainer(mocks.PodNotPresent)
		Expect(container).NotTo(BeNil())

		var isResolved bool
		isResolved, err = container.IsResolved(context.TODO(), testEntrypoint)
		Expect(isResolved).To(BeFalse())
		Expect(err).ToNot(HaveOccurred())
	})
})
