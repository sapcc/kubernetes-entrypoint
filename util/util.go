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

package util

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"opendev.org/airship/kubernetes-entrypoint/logger"
	"opendev.org/airship/kubernetes-entrypoint/util/env"
)

func GetIp() (string, error) {
	var iface string
	if iface = os.Getenv("INTERFACE_NAME"); iface == "" {
		return "", errors.New("environment variable INTERFACE_NAME not set")
	}
	i, err := net.InterfaceByName(iface)
	if err != nil {
		return "", fmt.Errorf("cannot get iface: %w", err)
	}

	address, err := i.Addrs()
	if err != nil || len(address) == 0 {
		return "", fmt.Errorf("cannot get ip: %w", err)
	}
	// Take first element to get rid of subnet
	ip := strings.Split(address[0].String(), "/")[0]
	return ip, nil
}

func ContainsSeparator(envString, kind string) bool {
	if strings.Contains(envString, env.Separator) {
		logger.Error.Printf("%s doesn't accept namespace: %s", kind, envString)
		return true
	}
	return false
}
