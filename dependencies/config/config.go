package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	entry "opendev.org/airship/kubernetes-entrypoint/entrypoint"
	"opendev.org/airship/kubernetes-entrypoint/logger"
	"opendev.org/airship/kubernetes-entrypoint/util"
	"opendev.org/airship/kubernetes-entrypoint/util/env"
)

const (
	configmapDirPrefix    = "/configmaps"
	NamespaceNotSupported = "Config doesn't accept namespace"
)

type configParams struct {
	HOSTNAME  string
	IP        string
	IP_ERLANG string
}

type Config struct {
	name   string
	params configParams
	prefix string
}

func init() {
	configEnv := entry.DependencyPrefix + "CONFIG"
	if util.ContainsSeparator(configEnv, "Config") {
		logger.Error.Print(NamespaceNotSupported)
		os.Exit(1)
	}
	if configDeps := env.SplitEnvToDeps(configEnv); len(configDeps) > 0 {
		for _, dep := range configDeps {
			config, err := NewConfig(dep.Name, configmapDirPrefix)
			if err != nil {
				logger.Error.Printf("Cannot initialize config dep: %v", err)
			}
			entry.Register(config)
		}
	}
}

func NewConfig(name string, prefix string) (*Config, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("cannot determine hostname: %w", err)
	}

	ip, err := util.GetIp()
	if err != nil {
		return nil, fmt.Errorf("cannot get ip address: %w", err)
	}

	return &Config{
		name: name,
		params: configParams{
			IP:        ip,
			IP_ERLANG: strings.ReplaceAll(ip, ".", ","),
			HOSTNAME:  hostname,
		},
		prefix: prefix,
	}, nil
}

func (c Config) IsResolved(ctx context.Context, entrypoint entry.EntrypointInterface) (bool, error) {
	// Create directory to ensure it exists
	if err := createDirectory(c.name); err != nil {
		return false, fmt.Errorf("couldn't create directory: %w", err)
	}
	if err := c.createAndTemplateConfig(); err != nil {
		return false, fmt.Errorf("cannot template %s: %w", c.name, err)
	}
	return true, nil
}

func (c Config) createAndTemplateConfig() error {
	config, err := os.Create(c.name)
	if err != nil {
		return err
	}
	file := filepath.Base(c.name)

	temp := template.Must(template.New(file).ParseFiles(getSrcConfig(c.prefix, file)))
	if err = temp.Execute(config, c.params); err != nil {
		return err
	}
	return nil
}

func getSrcConfig(prefix string, config string) (srcConfig string) {
	return fmt.Sprintf("%s/%s/%s", prefix, config, config)
}

func createDirectory(file string) error {
	return os.MkdirAll(filepath.Dir(file), 0o755)
}

func (c Config) String() string {
	return "Config " + c.name
}
