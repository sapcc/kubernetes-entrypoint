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

package pod

import (
	"context"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"opendev.org/airship/kubernetes-entrypoint/entrypoint"
	"opendev.org/airship/kubernetes-entrypoint/mocks"
)

const (
	podEnvVariableValue = "podlist"
	podNamespace        = "test"
	requireSameNode     = true
)

var (
	testEntrypoint entrypoint.EntrypointInterface
	testLabels     = map[string]string{"foo": "bar"}
)

var _ = Describe("Pod", func() {
	BeforeEach(func() {
		err := os.Setenv(PodNameEnvVar, podEnvVariableValue)
		Expect(err).NotTo(HaveOccurred())

		testEntrypoint = mocks.NewEntrypoint()
	})

	It(fmt.Sprintf("checks failure of new pod creation without %s set", PodNameEnvVar), func() {
		os.Unsetenv(PodNameEnvVar)
		pod, err := NewPod(testLabels, podNamespace, requireSameNode)

		Expect(pod).To(BeNil())
		Expect(err.Error()).To(Equal(fmt.Sprintf(PodNameNotSetErrorFormat, podNamespace)))
	})

	It(fmt.Sprintf("creates new pod with %s set and checks its name", PodNameEnvVar), func() {
		pod, err := NewPod(testLabels, podNamespace, requireSameNode)
		Expect(pod).NotTo(BeNil())
		Expect(err).NotTo(HaveOccurred())

		Expect(pod.labels).To(Equal(testLabels))
	})

	It("is resolved via all pods matching labels ready on same host", func() {
		pod, err := NewPod(map[string]string{"name": mocks.SameHostReadyMatchLabel}, podNamespace, requireSameNode)
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := pod.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
	})

	It("is resolved via some pods matching labels ready on same host", func() {
		pod, err := NewPod(map[string]string{"name": mocks.SameHostSomeReadyMatchLabel}, podNamespace, requireSameNode)
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := pod.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
	})

	It("is not resolved via a pod matching labels not ready on same host", func() {
		pod, err := NewPod(map[string]string{"name": mocks.SameHostNotReadyMatchLabel}, podNamespace, requireSameNode)
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := pod.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeFalse())
		Expect(err).To(HaveOccurred())
	})

	It("is not resolved via pod matching labels ready on different host", func() {
		pod, err := NewPod(map[string]string{"name": mocks.DifferentHostReadyMatchLabel}, podNamespace, requireSameNode)
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := pod.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeFalse())
		Expect(err).To(HaveOccurred())
	})

	It("is resolved via pod matching labels ready on different host when requireSameNode=false", func() {
		pod, err := NewPod(map[string]string{"name": mocks.DifferentHostReadyMatchLabel}, podNamespace, false)
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := pod.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
	})

	It("is not resolved via pod matching labels not ready on different host when requireSameNode=false", func() {
		pod, err := NewPod(map[string]string{"name": mocks.DifferentHostNotReadyMatchLabel}, podNamespace, false)
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := pod.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeFalse())
		Expect(err).To(HaveOccurred())
	})

	It("is not resolved via no pods matching labels", func() {
		pod, err := NewPod(map[string]string{"name": mocks.NoPodsMatchLabel}, podNamespace, requireSameNode)
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := pod.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeFalse())
		Expect(err).To(HaveOccurred())
	})

	It("is not resolved when getting pods matching labels from api fails", func() {
		pod, err := NewPod(map[string]string{"name": mocks.FailingMatchLabel}, podNamespace, requireSameNode)
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := pod.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeFalse())
		Expect(err).To(HaveOccurred())
	})

	It(fmt.Sprintf("is not resolved when getting current pod via %s value fails", PodNameEnvVar), func() {
		os.Setenv(PodNameEnvVar, mocks.PodNotPresent)
		pod, err := NewPod(map[string]string{"name": mocks.SameHostReadyMatchLabel}, podNamespace, requireSameNode)
		Expect(err).NotTo(HaveOccurred())

		isResolved, err := pod.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeFalse())
		Expect(err).To(HaveOccurred())
	})
})
