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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"opendev.org/airship/kubernetes-entrypoint/entrypoint"
	"opendev.org/airship/kubernetes-entrypoint/mocks"
)

const (
	testJobName      = "TEST_JOB_NAME"
	testJobNamespace = "TEST_JOB_NAMESPACE"
)

var testLabels = map[string]string{
	"k1": "v1",
}

var testEntrypoint entrypoint.EntrypointInterface

var _ = Describe("Job", func() {
	BeforeEach(func() {
		testEntrypoint = mocks.NewEntrypoint()
	})

	It("constructor correctly assigns fields", func() {
		nameJob := NewJob(testJobName, testJobNamespace, nil)

		Expect(nameJob.name).To(Equal(testJobName))
		Expect(nameJob.namespace).To(Equal(testJobNamespace))

		labelsJob := NewJob("", testJobNamespace, testLabels)

		Expect(labelsJob.labels).To(Equal(testLabels))
	})

	It("constructor returns nil when both name and labels specified", func() {
		job := NewJob(testJobName, testJobNamespace, testLabels)

		Expect(job).To(BeNil())
	})

	It("checks resolution of a succeeding job by name", func() {
		job := NewJob(mocks.SucceedingJobName, mocks.SucceedingJobName, nil)

		isResolved, err := job.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
	})

	It("checks resolution failure of a failing job by name", func() {
		job := NewJob(mocks.FailingJobName, mocks.FailingJobName, nil)

		isResolved, err := job.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeFalse())
		Expect(err.Error()).To(Equal(fmt.Sprintf(FailingStatusFormat, job)))
	})

	It("checks resolution of a succeeding job by labels", func() {
		job := NewJob("", mocks.SucceedingJobName, map[string]string{"name": mocks.SucceedingJobLabel})

		isResolved, err := job.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
	})

	It("checks resolution failure of a failing job by labels", func() {
		job := NewJob("", mocks.FailingJobName, map[string]string{"name": mocks.FailingJobLabel})

		isResolved, err := job.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeFalse())
		Expect(err.Error()).To(Equal(fmt.Sprintf(FailingStatusFormat, job)))
	})
})
