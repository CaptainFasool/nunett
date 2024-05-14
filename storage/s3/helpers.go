package s3

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// GetAWSDefaultConfig returns the default AWS config based on environment variables,
// shared configuration and shared credentials files.
func GetAWSDefaultConfig() (aws.Config, error) {
	var optFns []func(*config.LoadOptions) error
	return config.LoadDefaultConfig(context.Background(), optFns...)
}

func hasValidCredentials(config aws.Config) bool {
	credentials, err := config.Credentials.Retrieve(context.Background())
	if err != nil {
		return false
	}
	return credentials.HasKeys()
}

// sanitizeKey removes trailing spaces and wildcards
func sanitizeKey(key string) string {
	return strings.TrimSuffix(strings.TrimSpace(key), "*")
}
