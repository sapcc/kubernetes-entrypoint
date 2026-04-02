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

package daemonset

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
	podEnvVariableValue = "podlist"
	daemonsetNamespace  = "test"
)

var testEntrypoint entrypoint.EntrypointInterface

var _ = Describe("Daemonset", func() {
	BeforeEach(func() {
		err := os.Setenv(PodNameEnvVar, podEnvVariableValue)
		Expect(err).NotTo(HaveOccurred())

		testEntrypoint = mocks.NewEntrypoint()
	})

	It(fmt.Sprintf("checks failure of new daemonset creation without %s set", PodNameEnvVar), func() {
		os.Unsetenv(PodNameEnvVar)
		daemonset, err := NewDaemonset(mocks.SucceedingDaemonsetName, daemonsetNamespace)

		Expect(daemonset).To(BeNil())
		errMsg := fmt.Sprintf(PodNameNotSetErrorFormat, mocks.SucceedingDaemonsetName, daemonsetNamespace)
		Expect(err.Error()).To(Equal(errMsg))
	})

	It(fmt.Sprintf("creates new daemonset with %s set and checks its name", PodNameEnvVar), func() {
		daemonset, err := NewDaemonset(mocks.SucceedingDaemonsetName, daemonsetNamespace)
		Expect(daemonset).NotTo(BeNil())
		Expect(err).NotTo(HaveOccurred())

		Expect(daemonset.name).To(Equal(mocks.SucceedingDaemonsetName))
	})

	It("checks resolution of a succeeding daemonset", func() {
		daemonset, err := NewDaemonset(mocks.SucceedingDaemonsetName, daemonsetNamespace)
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := daemonset.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
	})

	It("checks resolution failure of a daemonset with incorrect name", func() {
		daemonset, err := NewDaemonset(mocks.FailingDaemonsetName, daemonsetNamespace)
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := daemonset.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeFalse())
		Expect(err).To(HaveOccurred())
	})

	It("checks resolution failure of a daemonset with incorrect match labels", func() {
		daemonset, err := NewDaemonset(mocks.FailingMatchLabelsDaemonsetName, daemonsetNamespace)
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := daemonset.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeFalse())
		Expect(err).To(HaveOccurred())
	})

	It(fmt.Sprintf("checks resolution failure of a daemonset with incorrect %s value", PodNameEnvVar), func() {
		// Set POD_NAME to value not present in the mocks
		os.Setenv(PodNameEnvVar, mocks.PodNotPresent)
		daemonset, err := NewDaemonset(mocks.FailingMatchLabelsDaemonsetName, daemonsetNamespace)
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := daemonset.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeFalse())
		Expect(err).To(HaveOccurred())
	})

	It("checks resolution failure of a daemonset with none of the pods with Ready status", func() {
		daemonset, err := NewDaemonset(mocks.NotReadyMatchLabelsDaemonsetName, daemonsetNamespace)
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := daemonset.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeFalse())
		Expect(err).To(HaveOccurred())
	})

	It("checks resolution of a correct daemonset namespace", func() {
		daemonset, err := NewDaemonset(mocks.CorrectNamespaceDaemonsetName, daemonsetNamespace)

		Expect(daemonset).NotTo(BeNil())
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := daemonset.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
	})

	It("checks resolution of an incorrect daemonset namespace", func() {
		daemonset, err := NewDaemonset(mocks.IncorrectNamespaceDaemonsetName, daemonsetNamespace)

		Expect(daemonset).NotTo(BeNil())
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := daemonset.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeFalse())
		Expect(err).To(HaveOccurred())
	})

	It("resolve daemonset and entrypoint pod in different namespaces", func() {
		daemonset, err := NewDaemonset(mocks.CorrectNamespaceDaemonsetName, mocks.CorrectDaemonsetNamespace)
		Expect(err).NotTo(HaveOccurred())

		err = os.Setenv(PodNameEnvVar, "shouldwork")
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := daemonset.IsResolved(context.TODO(), testEntrypoint)

		Expect(err).NotTo(HaveOccurred())
		Expect(isResolved).To(BeTrue())
	})
})
