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

package socket

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"opendev.org/airship/kubernetes-entrypoint/entrypoint"
	"opendev.org/airship/kubernetes-entrypoint/mocks"
)

const (
	tempPathSuffix = "k8s-entrypoint"

	existingSocket    = "existing-socket"
	nonExistingSocket = "nonexisting-socket"
)

var (
	testDir string

	existingSocketPath    string
	nonExistingSocketPath string
)

var testEntrypoint entrypoint.EntrypointInterface

var _ = Describe("Socket", func() {
	// NOTE: It is impossible for a user to create a file that he does not
	// have access to, and thus it is impossible to write an isolated unit
	// test that checks for permission errors. That test is omitted from
	// this suite

	BeforeEach(func() {
		testEntrypoint = mocks.NewEntrypoint()

		var err error
		testDir, err = os.MkdirTemp("", tempPathSuffix)
		Expect(err).NotTo(HaveOccurred())

		existingSocketPath = filepath.Join(testDir, existingSocket)
		nonExistingSocketPath = filepath.Join(testDir, nonExistingSocket)

		_, err = os.Create(existingSocketPath)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(testDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("checks the name of a newly created socket", func() {
		socket := NewSocket(existingSocketPath)

		Expect(socket.name).To(Equal(existingSocketPath))
	})

	It("resolves an existing socket", func() {
		socket := NewSocket(existingSocketPath)

		isResolved, err := socket.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
	})

	It("fails on trying to resolve a nonexisting socket", func() {
		socket := NewSocket(nonExistingSocketPath)

		isResolved, err := socket.IsResolved(context.TODO(), testEntrypoint)

		Expect(isResolved).To(BeFalse())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(fmt.Sprintf(NonExistingErrorFormat, socket)))
	})
})
