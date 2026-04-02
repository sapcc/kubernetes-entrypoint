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

package entrypoint

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cli "opendev.org/airship/kubernetes-entrypoint/client"
	"opendev.org/airship/kubernetes-entrypoint/logger"
	"opendev.org/airship/kubernetes-entrypoint/mocks"
)

const (
	testNamespace     = "test"
	dummyResolverName = "dummy"
	loggerInfoText    = "Entrypoint INFO: "
)

var (
	testEntrypoint EntrypointInterface
	testClient     cli.ClientInterface
)

type dummyResolver struct {
	name      string
	namespace string
}

func (d dummyResolver) IsResolved(ctx context.Context, entry EntrypointInterface) (bool, error) {
	return true, nil
}

func (d dummyResolver) GetName() (name string) {
	return d.name
}

func (d dummyResolver) String() string {
	return fmt.Sprintf("Dummy %s in namespace %s", d.name, d.namespace)
}

func init() {
	testClient = mocks.NewClient()
	testEntrypoint = mocks.NewEntrypointInNamespace(testNamespace)
}

func registerNilResolver() {
	Register(nil)
}

var _ = Describe("Entrypoint", func() {
	dummy := dummyResolver{name: dummyResolverName}

	BeforeEach(func() {
		logger.Info.SetFlags(0)
		logger.Warning.SetFlags(0)
		logger.Error.SetFlags(0)
	})

	AfterEach(func() {
		// Clear dependencies
		dependencies = make([]Resolver, 0)
	})

	It("registers new nil resolver", func() {
		defer GinkgoRecover()

		Ω(registerNilResolver).Should(Panic())
	})

	It("registers new non-nil resolver", func() {
		defer GinkgoRecover()
		Register(dummy)
		Expect(dependencies).To(HaveLen(1))
	})

	It("checks Client() method", func() {
		client := testEntrypoint.Client()
		Expect(client).To(Equal(testClient))
	})

	It("resolves main entrypoint with a dummy dependency", func() {
		defer GinkgoRecover()

		// Set output logger to our reader
		r, w, err := os.Pipe()
		Expect(err).NotTo(HaveOccurred())
		tmp := os.Stdout
		defer func() {
			os.Stdout = tmp
		}()

		logger.Info.SetOutput(w)

		os.Stdout = w
		go func() {
			mainEntrypoint := Entrypoint{client: mocks.NewClient(), namespace: "main"}
			Register(dummy)
			mainEntrypoint.Resolve()
			w.Close()
		}()

		// Wait for resolver to finish
		time.Sleep(5 * time.Second)

		stdout, err := io.ReadAll(r)
		Expect(err).NotTo(HaveOccurred())
		resolvedString := fmt.Sprintf("%sResolving %v\n%sDependency %v is resolved.\n",
			loggerInfoText, dummy, loggerInfoText, dummy)
		Expect(string(stdout)).To(Equal(resolvedString))
	})
})
