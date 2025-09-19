# OP-Viper

A Go library that integrates 1Password with Viper configuration management, allowing you to securely inject secrets from 1Password into your application configuration.

## Installation

```bash
go get github.com/rafaelbroseghini/op-viper
```

## Usage

### Basic Setup

```go
package main

import (
    "context"
    "github.com/spf13/viper"
    "github.com/rafaelbroseghini/op-viper/pkg/onepassword"
)

func main() {
    v := viper.New()
    v.SetConfigType("yaml")
    v.SetConfigName("config.yaml")
    v.AddConfigPath(".")
    
    v.ReadInConfig()
    
    var config YourConfig
    ctx := context.Background()
    
    // Using CLI client
    cliClient := &onepassword.OnePasswordCLIClient{}
    v.Unmarshal(&config, viper.DecoderConfigOption(onepassword.OnePasswordHookFunc(ctx, cliClient)))
}
```

### Using 1Password SDK Client

```go
// Initialize SDK client with service account
ctx := context.Background()
sdkClient := onepassword.NewOnePasswordSDKClient(
    ctx,
    "my-app",           // integration name
    "1.0.0",           // version
    "your-service-account-token",
)

v.Unmarshal(&config, viper.DecoderConfigOption(onepassword.OnePasswordHookFunc(ctx, sdkClient)))
```

### Configuration File Format

In your YAML configuration file, reference 1Password secrets using the following syntax:

```yaml
database:
  host: "localhost"
  port: 5432
  username: "{{ op://vault/item/username }}"
  password: "{{ op://vault/item/password }}"

api:
  key: "{{ op://production/api-key }}"
  secret: "{{ op://production/api-secret }}"
```

The library will automatically:
1. Detect strings wrapped in `{{ }}`
2. Extract the 1Password reference (must start with `op://`)
3. Resolve the secret using your configured client
4. Replace the placeholder with the actual secret value

## License

This project is licensed under the MIT License.
