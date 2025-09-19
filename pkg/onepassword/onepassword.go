package onepassword

import (
	"context"
	"os/exec"
	"reflect"
	"strings"

	op "github.com/1password/onepassword-sdk-go"
	"github.com/go-viper/mapstructure/v2"
)

type OnePasswordClient interface {
	Get(ctx context.Context, path string) (string, error)
}

type OnePasswordCLIClient struct{}

func (c OnePasswordCLIClient) Get(ctx context.Context, path string) (string, error) {
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

func (c OnePasswordSDKClient) Get(ctx context.Context, path string) (string, error) {
	item, err := c.Client.Secrets().Resolve(ctx, path)
	if err != nil {
		return "", err
	}
	return item, nil
}

func OnePasswordHookFunc(ctx context.Context, c OnePasswordClient) mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		value := data.(string)
		if value == "" {
			return data, nil
		}

		if !strings.HasPrefix(value, "{{") || !strings.HasSuffix(value, "}}") {
			return value, nil
		}

		value = strings.TrimPrefix(value, "{{")
		value = strings.TrimSuffix(value, "}}")
		value = strings.TrimSpace(value)

		if !strings.HasPrefix(value, "op://") {
			return value, nil
		}

		return c.Get(ctx, value)
	}
}
