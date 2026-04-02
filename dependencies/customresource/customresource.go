package customresource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"k8s.io/client-go/util/jsonpath"

	"opendev.org/airship/kubernetes-entrypoint/entrypoint"
	"opendev.org/airship/kubernetes-entrypoint/logger"
	"opendev.org/airship/kubernetes-entrypoint/util/env"
)

// A Resolver represents the state of a CustomResource
type Resolver struct {
	APIVersion string  `json:"apiVersion"`
	Kind       string  `json:"kind"`
	Name       string  `json:"name"`
	Namespace  string  `json:"namespace"`
	Fields     []Field `json:"fields"`
}

var _ entrypoint.Resolver = Resolver{}

// A Field represents a key-value pair
type Field struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func init() {
	crEnv := entrypoint.DependencyPrefix + "CUSTOM_RESOURCE"
	resolvers, err := fromEnv(crEnv)
	if err != nil {
		logger.Error.Printf("Error initializing custom resource: %s", err.Error()) // Fixed format string
	}
	for _, resolver := range resolvers {
		entrypoint.Register(resolver)
	}
}

// IsResolved will return true when the values for each key in r.Fields is the same as the resource in the cluster
func (r Resolver) IsResolved(ctx context.Context, ep entrypoint.EntrypointInterface) (bool, error) {
	customResource, err := ep.Client().CustomResource(ctx, r.APIVersion, r.Kind, r.Namespace, r.Name)
	if err != nil {
		return false, err
	}

	for _, field := range r.Fields {
		expected := field.Value

		// Parse the key into a JSONPath object (handles dot-notation, $-prefix, and {}-wrapped formats)
		jp, err := parseKey(field.Key)
		if err != nil {
			return false, err
		}

		// Evaluate the JSONPath to extract the value
		actual, err := evaluateJSONPath(jp, customResource.Object)
		if err != nil {
			return false, err
		}

		if actual == expected {
			continue
		}

		// Convert actual value to JSON string for comparison to ensure consistent representation
		actualBytes, err := json.Marshal(actual)
		if err != nil {
			return false, fmt.Errorf("failed to marshal value for key [%s]: %w", field.Key, err)
		}

		actualStr := string(actualBytes)
		if actualStr != expected {
			return false, fmt.Errorf("expected value of [%s] to be [%s], but got [%s]", field.Key, expected, actualStr)
		}
	}

	return true, nil
}

// parseKey normalizes a key (dot-notation, $-prefixed, or {}-wrapped JSONPath) and returns a parsed JSONPath object.
func parseKey(key string) (*jsonpath.JSONPath, error) {
	// Normalize to JSONPath format with curly braces
	normalizedKey := key
	if !strings.HasPrefix(key, "{") {
		if strings.HasPrefix(key, "$") {
			normalizedKey = "{" + key + "}"
		} else {
			// Dot-notation, convert to JSONPath
			normalizedKey = "{$." + key + "}"
		}
	}

	// Create and parse the JSONPath
	jp := jsonpath.New("resolver")
	if err := jp.Parse(normalizedKey); err != nil {
		return nil, fmt.Errorf("invalid key [%s]: %w", key, err)
	}

	return jp, nil
}

// evaluateJSONPath executes a parsed JSONPath against a resource and returns the value.
// It returns an error if the JSONPath returns no results, multiple results, or invalid values.
func evaluateJSONPath(jp *jsonpath.JSONPath, resource map[string]any) (any, error) {
	results, err := jp.FindResults(resource)
	if err != nil {
		return nil, fmt.Errorf("error executing JSONPath: %w", err)
	}

	// Ensure we have exactly one result set
	if len(results) == 0 || len(results[0]) == 0 {
		return nil, errors.New("JSONPath returned no results")
	}

	if len(results) > 1 || len(results[0]) > 1 {
		return nil, errors.New("JSONPath returned multiple values; only single-value results are allowed")
	}

	// Get the single result value
	value := results[0][0]

	// Extract the actual value from reflect.Value
	if !value.IsValid() {
		return nil, errors.New("JSONPath returned invalid value")
	}

	return value.Interface(), nil
}

// fromEnv reads the value of the jsonEnv variable and returns the array of
// Resolvers it contains, if any
func fromEnv(jsonEnv string) ([]Resolver, error) {
	resolvers := []Resolver{}
	jsonEnvVal, isSet := os.LookupEnv(jsonEnv)
	if !isSet {
		return resolvers, nil
	}

	// Check if the environment variable is empty
	if strings.TrimSpace(jsonEnvVal) == "" {
		return resolvers, nil
	}

	err := json.Unmarshal([]byte(jsonEnvVal), &resolvers)
	if err != nil {
		return resolvers, fmt.Errorf("unable to unmarshal variable %s with value %s: %s",
			jsonEnv, jsonEnvVal, err.Error())
	}

	namespace := env.GetBaseNamespace()
	for i := range resolvers {
		if resolvers[i].Namespace == "" {
			resolvers[i].Namespace = namespace
		}
	}

	return resolvers, nil
}
