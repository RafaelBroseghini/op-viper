package onepassword

import (
	"context"
	"os/exec"
	"reflect"
	"strings"

	op "github.com/1password/onepassword-sdk-go"
	"github.com/go-viper/mapstructure/v2"
)

const (
	OpenCurlyBrace  = "{{"
	CloseCurlyBrace = "}}"
)

type Loader struct {
	Prefix string
	Suffix string
	Client OnePasswordClient
}

type LoaderFunc func(h *Loader)

func NewDefaultLoader() Loader {
	return Loader{Prefix: OpenCurlyBrace, Suffix: CloseCurlyBrace, Client: OnePasswordCLIClient{}}
}

func NewLoader(opts ...LoaderFunc) Loader {
	h := Loader{}
	for _, opt := range opts {
		opt(&h)
	}
	return h
}

func WithPrefix(prefix string) LoaderFunc {
	return func(h *Loader) {
		h.Prefix = prefix
	}
}

func WithSuffix(suffix string) LoaderFunc {
	return func(h *Loader) {
		h.Suffix = suffix
	}
}

func WithCLIClient(client OnePasswordCLIClient) LoaderFunc {
	return func(h *Loader) {
		h.Client = client
	}
}

func WithSDKClient(client OnePasswordSDKClient) LoaderFunc {
	return func(h *Loader) {
		h.Client = client
	}
}

type OnePasswordClient interface {
	Resolve(ctx context.Context, path string) (string, error)
}

type OnePasswordCLIClient struct{}

func (c OnePasswordCLIClient) Resolve(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "op", "read", path)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

type OnePasswordSDKClient struct {
	Client *op.Client
}

func NewOnePasswordSDKClient(ctx context.Context, name, version, token string) OnePasswordSDKClient {
	client, err := op.NewClient(ctx, op.WithServiceAccountToken(token), op.WithIntegrationInfo(name, version))
	if err != nil {
		panic(err)
	}
	return OnePasswordSDKClient{Client: client}
}

func (c OnePasswordSDKClient) Resolve(ctx context.Context, path string) (string, error) {
	item, err := c.Client.Secrets().Resolve(ctx, path)
	if err != nil {
		return "", err
	}
	return item, nil
}

func (h Loader) OnePasswordHookFunc(ctx context.Context) mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		value := data.(string)
		if value == "" {
			return data, nil
		}

		if !strings.HasPrefix(value, h.Prefix) || !strings.HasSuffix(value, h.Suffix) {
			return value, nil
		}

		value = strings.TrimPrefix(value, h.Prefix)
		value = strings.TrimSuffix(value, h.Suffix)
		value = strings.TrimSpace(value)

		if !strings.HasPrefix(value, "op://") {
			return value, nil
		}

		return h.Client.Resolve(ctx, value)
	}
}
