package onepassword

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type MockOnePasswordClient struct {
	ResolveFunc func(ctx context.Context, path string) (string, error)
}

func (m *MockOnePasswordClient) Resolve(ctx context.Context, path string) (string, error) {
	if m.ResolveFunc != nil {
		return m.ResolveFunc(ctx, path)
	}
	return "", errors.New("mock not configured")
}

func TestNewDefaultLoader(t *testing.T) {
	loader := NewDefaultLoader()

	if loader.Prefix != OpenCurlyBrace {
		t.Errorf("Expected prefix %s, got %s", OpenCurlyBrace, loader.Prefix)
	}

	if loader.Suffix != CloseCurlyBrace {
		t.Errorf("Expected suffix %s, got %s", CloseCurlyBrace, loader.Suffix)
	}

	if loader.Client == nil {
		t.Error("Expected client to be set")
	}

	if _, ok := loader.Client.(OnePasswordCLIClient); !ok {
		t.Error("Expected default client to be OnePasswordCLIClient")
	}
}

func TestNewLoader(t *testing.T) {
	tests := []struct {
		name     string
		opts     []LoaderFunc
		expected Loader
	}{
		{
			name: "empty options",
			opts: []LoaderFunc{},
			expected: Loader{
				Prefix: "",
				Suffix: "",
				Client: nil,
			},
		},
		{
			name: "with prefix",
			opts: []LoaderFunc{WithPrefix("${")},
			expected: Loader{
				Prefix: "${",
				Suffix: "",
				Client: nil,
			},
		},
		{
			name: "with suffix",
			opts: []LoaderFunc{WithSuffix("}")},
			expected: Loader{
				Prefix: "",
				Suffix: "}",
				Client: nil,
			},
		},
		{
			name: "with CLI client",
			opts: []LoaderFunc{WithCLIClient(OnePasswordCLIClient{})},
			expected: Loader{
				Prefix: "",
				Suffix: "",
				Client: OnePasswordCLIClient{},
			},
		},
		{
			name: "multiple options",
			opts: []LoaderFunc{
				WithPrefix("${"),
				WithSuffix("}"),
				WithCLIClient(OnePasswordCLIClient{}),
			},
			expected: Loader{
				Prefix: "${",
				Suffix: "}",
				Client: OnePasswordCLIClient{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader(tt.opts...)

			if loader.Prefix != tt.expected.Prefix {
				t.Errorf("Expected prefix %s, got %s", tt.expected.Prefix, loader.Prefix)
			}

			if loader.Suffix != tt.expected.Suffix {
				t.Errorf("Expected suffix %s, got %s", tt.expected.Suffix, loader.Suffix)
			}

			if loader.Client != nil && tt.expected.Client == nil {
				t.Error("Expected client to be nil")
			}

			if loader.Client == nil && tt.expected.Client != nil {
				t.Error("Expected client to be set")
			}
		})
	}
}

func TestOnePasswordHookFunc(t *testing.T) {
	loader := NewLoader(func(l *Loader) {
		l.Client = &MockOnePasswordClient{
			ResolveFunc: func(ctx context.Context, path string) (string, error) {
				return "resolved", nil
			},
		}
	}, WithPrefix("{{"), WithSuffix("}}"))

	ctx := context.Background()

	hookFunc := loader.OnePasswordHookFunc(ctx)
	hookFuncTyped := hookFunc.(func(reflect.Type, reflect.Type, any) (any, error))
	processedValue, err := hookFuncTyped(reflect.TypeOf(""), reflect.TypeOf(""), "{{op://vault/item}}")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if processedValue != "resolved" {
		t.Errorf("Expected resolved, got %s", processedValue)
	}
}
