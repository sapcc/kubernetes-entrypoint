package customresource

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"opendev.org/airship/kubernetes-entrypoint/mocks"
)

const listOfTwoCRResolvers = `
[
  {
    "apiVersion": "api1",
    "kind": "kind1",
    "namespace": "foospace1",
    "name": "foo1",
    "fields": [
      {
        "key": "field1key1",
        "value": "field1val1"
      },
      {
        "key": "field1key2",
        "value": "field1val2"
      }
    ]
  },
  {
    "apiVersion": "api2",
    "kind": "kind2",
    "namespace": "foospace2",
    "name": "foo2",
    "fields": [
      {
        "key": "field2key1",
        "value": "field2val1"
      },
      {
        "key": "field2key2",
        "value": "field2val2"
      }
    ]
  }
]`

func TestFromEnv(t *testing.T) {
	tests := []struct {
		name      string
		useEnvVar bool
		envVar    string
		expected  []Resolver
		expectErr bool
	}{
		{
			name:      "EmptyVar",
			useEnvVar: false,
			envVar:    "",
			expected:  []Resolver{},
			expectErr: false,
		},
		{
			name:      "InvalidVar",
			useEnvVar: true,
			envVar:    `[}"invalid": "json"}]`,
			expected:  []Resolver{},
			expectErr: true,
		},
		{
			name:      "Successful",
			useEnvVar: true,
			envVar:    listOfTwoCRResolvers,
			expected: []Resolver{
				{
					APIVersion: "api1",
					Kind:       "kind1",
					Namespace:  "foospace1",
					Name:       "foo1",
					Fields: []Field{
						{
							Key:   "field1key1",
							Value: "field1val1",
						},
						{
							Key:   "field1key2",
							Value: "field1val2",
						},
					},
				},
				{
					APIVersion: "api2",
					Kind:       "kind2",
					Namespace:  "foospace2",
					Name:       "foo2",
					Fields: []Field{
						{
							Key:   "field2key1",
							Value: "field2val1",
						},
						{
							Key:   "field2key2",
							Value: "field2val2",
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name:      "UnNamespaced",
			useEnvVar: true,
			envVar:    `[{"apiVersion":"api","kind":"kind","name":"foo","fields":[]}]`,
			expected: []Resolver{
				{
					APIVersion: "api",
					Kind:       "kind",
					Namespace:  "default",
					Name:       "foo",
					Fields:     []Field{},
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.useEnvVar {
				os.Setenv("TEST_LIST_JSON", tt.envVar)
				defer os.Unsetenv("TEST_LIST_JSON")
			} else {
				os.Unsetenv("TEST_LIST_JSON")
			}

			actual, err := fromEnv("TEST_LIST_JSON")
			if err != nil {
				if !tt.expectErr {
					t.Fatalf("Unexpected error: %s", err.Error())
				}
			} else if tt.expectErr {
				t.Fatalf("Expected error, but received none")
			}

			if !reflect.DeepEqual(tt.expected, actual) {
				t.Errorf("Expected %+v, got %+v", tt.expected, actual)
			}
		})
	}
}

func TestIsResolved(t *testing.T) {
	tests := []struct {
		name           string
		customResource *unstructured.Unstructured
		resolver       Resolver
		expected       bool
		expectErr      bool
		clientErr      error
	}{
		{
			name: "Successful",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"simpleKey":  "ready",
					"complex": map[string]interface{}{
						"key": map[string]interface{}{
							"with": map[string]interface{}{
								"layers": "ready",
							},
						},
					},
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "simpleKey",
						Value: "ready",
					},
					{
						Key:   "complex.key.with.layers",
						Value: "ready",
					},
				},
			},
			expected:  true,
			expectErr: false,
			clientErr: nil,
		},
		{
			name: "DotNotationNumericValue",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"replicas":   int64(3),
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "replicas",
						Value: "3",
					},
				},
			},
			expected:  true,
			expectErr: false,
			clientErr: nil,
		},
		{
			name: "DotNotationBooleanValue",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"enabled":    true,
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "enabled",
						Value: "true",
					},
				},
			},
			expected:  true,
			expectErr: false,
			clientErr: nil,
		},
		{
			name: "DotNotationNullValue",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"status":     nil,
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "status",
						Value: "null",
					},
				},
			},
			expected:  true,
			expectErr: false,
			clientErr: nil,
		},
		{
			name: "DotNotationNestedNumeric",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"spec": map[string]interface{}{
						"replicas": int64(5),
					},
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "spec.replicas",
						Value: "5",
					},
				},
			},
			expected:  true,
			expectErr: false,
			clientErr: nil,
		},
		{
			name: "DotNotationFloatValue",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"version":    2.5,
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "version",
						Value: "2.5",
					},
				},
			},
			expected:  true,
			expectErr: false,
			clientErr: nil,
		},
		{
			name: "JSONPathSimple",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"status":     "ready",
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "$.status",
						Value: "ready",
					},
				},
			},
			expected:  true,
			expectErr: false,
			clientErr: nil,
		},
		{
			name: "JSONPathNested",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"status": "ready",
						},
					},
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "$.spec.template.status",
						Value: "ready",
					},
				},
			},
			expected:  true,
			expectErr: false,
			clientErr: nil,
		},
		{
			name: "JSONPathNumericValue",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"replicas":   int64(3),
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "$.replicas",
						Value: "3",
					},
				},
			},
			expected:  true,
			expectErr: false,
			clientErr: nil,
		},
		{
			name: "JSONPathBooleanValue",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"enabled":    true,
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "$.enabled",
						Value: "true",
					},
				},
			},
			expected:  true,
			expectErr: false,
			clientErr: nil,
		},
		{
			name: "JSONPathArrayIndexing",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"items": []interface{}{
						map[string]interface{}{
							"status": "ready",
						},
						map[string]interface{}{
							"status": "pending",
						},
					},
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "$.items[0].status",
						Value: "ready",
					},
				},
			},
			expected:  true,
			expectErr: false,
			clientErr: nil,
		},
		{
			name: "JSONPathMixedWithDotNotation",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"status":     "ready",
					"simpleKey":  "active",
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "$.status",
						Value: "ready",
					},
					{
						Key:   "simpleKey",
						Value: "active",
					},
				},
			},
			expected:  true,
			expectErr: false,
			clientErr: nil,
		},
		{
			name: "JSONPathInvalidSyntax",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"status":     "ready",
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "$.[invalid",
						Value: "ready",
					},
				},
			},
			expected:  false,
			expectErr: true,
			clientErr: nil,
		},
		{
			name: "JSONPathNoResults",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "$.nonexistent",
						Value: "ready",
					},
				},
			},
			expected:  false,
			expectErr: true,
			clientErr: nil,
		},
		{
			name: "JSONPathMultipleResults",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"items": []interface{}{
						map[string]interface{}{
							"status": "ready",
						},
						map[string]interface{}{
							"status": "pending",
						},
					},
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "$.items[*].status",
						Value: "ready",
					},
				},
			},
			expected:  false,
			expectErr: true,
			clientErr: nil,
		},
		{
			name: "JSONPathNonPrimitiveObject",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"complex": map[string]interface{}{
						"nested": "value",
					},
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "$.complex",
						Value: "value",
					},
				},
			},
			expected:  false,
			expectErr: true,
			clientErr: nil,
		},
		{
			name: "JSONPathNonPrimitiveArray",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"items": []interface{}{
						"item1",
						"item2",
					},
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "$.items",
						Value: "item1",
					},
				},
			},
			expected:  false,
			expectErr: true,
			clientErr: nil,
		},
		{
			name: "JSONPathValueMismatch",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"status":     "pending",
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "$.status",
						Value: "ready",
					},
				},
			},
			expected:  false,
			expectErr: true,
			clientErr: nil,
		},
		{
			name: "JSONPathFloatValue",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"version":    3.14,
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "$.version",
						Value: "3.14",
					},
				},
			},
			expected:  true,
			expectErr: false,
			clientErr: nil,
		},
		{
			name: "JSONPathNullValue",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"status":     nil,
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "$.status",
						Value: "null",
					},
				},
			},
			expected:  true,
			expectErr: false,
			clientErr: nil,
		},
		{
			name: "JSONPathWithCurlyBraces",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"status":     "ready",
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "{.status}",
						Value: "ready",
					},
				},
			},
			expected:  true,
			expectErr: false,
			clientErr: nil,
		},
		{
			name: "Unresolved",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"simpleKey":  "NOT-ready",
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "simpleKey",
						Value: "ready",
					},
				},
			},
			expected:  false,
			expectErr: true,
			clientErr: nil,
		},
		{
			name:           "ClientError",
			customResource: &unstructured.Unstructured{},
			resolver:       Resolver{},
			expected:       false,
			expectErr:      true,
			clientErr:      errors.New("Generic error"),
		},
		{
			name: "BadCustomResource",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
					"notAMap":    5,
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "notAMap.nothingHere",
						Value: "ready",
					},
				},
			},
			expected:  false,
			expectErr: true,
			clientErr: nil,
		},
		{
			name: "MissingKey",
			customResource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "stable.example.com/v1",
					"kind":       "Foo",
					"name":       "my-foo",
					"namespace":  "default",
				},
			},
			resolver: Resolver{
				APIVersion: "stable.exmaple.com/v1",
				Kind:       "Foo",
				Name:       "my-foo",
				Fields: []Field{
					{
						Key:   "missingKey",
						Value: "ready",
					},
				},
			},
			expected:  false,
			expectErr: true,
			clientErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := mocks.NewEntrypoint()
			ep.MockClient.FakeCustomResource = tt.customResource
			ep.MockClient.Err = tt.clientErr
			result, err := tt.resolver.IsResolved(context.TODO(), ep)
			if err != nil {
				if !tt.expectErr {
					t.Fatalf("Unexpected error: %s", err.Error())
				}
			} else if tt.expectErr {
				t.Fatalf("Expected error, but received none")
			}

			if result != tt.expected {
				t.Errorf("Expected success to be %v, but got %v", tt.expected, result)
			}
		})
	}
}
