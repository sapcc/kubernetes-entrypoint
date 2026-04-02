// Copyright 2018 The kubernetes-entrypoint Authors
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
	"os"
	"os/exec"
	"syscall"

	"opendev.org/airship/kubernetes-entrypoint/logger"
)

func Execute(command []string) error {
	path, err := exec.LookPath(command[0])
	if err != nil {
		logger.Error.Printf("Cannot find a binary %v : %v", command[0], err)
		return err
	}

	env := os.Environ()
	err = syscall.Exec(path, command, env)
	if err != nil {
		logger.Error.Printf("Executing command %v failed: %v", command, err)
		return err
	}
	return nil
}
