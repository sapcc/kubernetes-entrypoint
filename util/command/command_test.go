// Copyright 2016 The kubernetes-entrypoint Authors
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

package command

import (
	"fmt"
	"syscall"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Execute", func() {
	AfterEach(func() {
		// Restore the real syscall after each spec.
		execSyscall = syscall.Exec
	})

	It("executes commands successfully", func() {
		execSyscall = func(argv0 string, argv []string, envv []string) error {
			return nil
		}
		err := Execute([]string{"echo", "test"})
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns an error when syscall.Exec fails", func() {
		execSyscall = func(_ string, _ []string, _ []string) error {
			return fmt.Errorf("exec failed")
		}
		err := Execute([]string{"echo", "test"})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error when command is not found", func() {
		err := Execute([]string{"nonexistent_command_xyz_abc"})
		Expect(err).To(HaveOccurred())
	})
})
