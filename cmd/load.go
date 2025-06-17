package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	yamlFile   string
	prefix     string
	secure     bool
	autoSecure bool
	overwrite  bool
)

var loadCmd = &cobra.Command{
	Use:     "load",
	Short:   "Read a YAML file to AWS SSM Parameter Store",
	Aliases: []string{"l", "lo"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if yamlFile == "" || prefix == "" {
			return fmt.Errorf("both --file and --prefix are required")
		}

		rawYaml, err := os.ReadFile(yamlFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		var data map[string]interface{}
		decoder := yaml.NewDecoder(bytes.NewReader(rawYaml))
		if err := decoder.Decode(&data); err != nil {
			return fmt.Errorf("failed to parse YAML: %w", err)
		}

		//awsCfg, err := config.LoadDefaultConfig(context.TODO())
		var cfgOpts []func(*config.LoadOptions) error
		if awsRegion != "" {
			cfgOpts = append(cfgOpts, config.WithRegion(awsRegion))
		}
		awsCfg, err := config.LoadDefaultConfig(ctx, cfgOpts...)
		if err != nil {
			return fmt.Errorf("failed to load AWS config: %w", err)
		}

		client := ssm.NewFromConfig(awsCfg)
		return loadConfig(data, prefix, client)
	},
}

func init() {
	loadCmd.Flags().StringVarP(&yamlFile, "file", "f", "", "Path to YAML config file (required)")
	loadCmd.Flags().StringVarP(&prefix, "prefix", "p", "", "SSM path prefix (e.g., /myapp) (required)")
	loadCmd.Flags().BoolVarP(&secure, "secure", "s", false, "Upload all parameters as SecureString")
	loadCmd.Flags().BoolVarP(&autoSecure, "auto-secure", "a", false, "Auto select SecureString for secret-like keys")
	loadCmd.Flags().BoolVarP(&showValues, "values", "v", false, "Show values while uploading")
	loadCmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "Allow overwriting existing parameters")
}

func loadConfig(cfg interface{}, path string, client *ssm.Client) error {
	switch val := cfg.(type) {
	case map[string]interface{}:
		for key, v := range val {
			childPath := strings.TrimSuffix(path, "/") + "/" + key
			if err := loadConfig(v, childPath, client); err != nil {
				return err
			}
		}
	case []interface{}:
		for i, v := range val {
			childPath := fmt.Sprintf("%s/%d", path, i)
			if err := loadConfig(v, childPath, client); err != nil {
				return err
			}
		}
	default:
		valueStr := fmt.Sprintf("%v", val)
		paramType := types.ParameterTypeString
		lockIcon := ""

		if secure {
			paramType = types.ParameterTypeSecureString
			lockIcon = " 🔒"
		} else if autoSecure && isSensitiveKey(path) {
			paramType = types.ParameterTypeSecureString
			lockIcon = " 🔒"
		}

		if showValues {
			fmt.Printf("Uploading %s%s = %s\n", path, lockIcon, valueStr)
		} else {
			fmt.Printf("Uploading %s%s\n", path, lockIcon)
		}
		_, err := client.PutParameter(ctx, &ssm.PutParameterInput{
			Name:      aws.String(path),
			Value:     aws.String(valueStr),
			Type:      paramType,
			Overwrite: aws.Bool(overwrite),
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to upload %s: %v\n", color.New(color.FgWhite, color.Bold).Sprint(path), color.New(color.FgRed).Sprint(extractMessage(err)))
		}
		return nil
		//return err
	}
	return nil
}

func isSensitiveKey(key string) bool {
	sensitive := []string{"password", "secret", "token", "key", "apikey", "auth", "private"}
	key = strings.ToLower(key)
	for _, s := range sensitive {
		if strings.Contains(key, s) {
			return true
		}
	}
	return false
}
