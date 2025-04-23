package cmd

import (
	"bufio"
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
	deleteFile   string
	deletePrefix string
	deleteYes    bool
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete parameters from AWS SSM based on a YAML file",
	Aliases: []string{"d", "de"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if deleteFile == "" || deletePrefix == "" {
			return fmt.Errorf("--file and --prefix are required")
		}

		rawYaml, err := os.ReadFile(deleteFile)
		if err != nil {
			return fmt.Errorf("error reading YAML file: %w", err)
		}

		var data map[string]interface{}
		decoder := yaml.NewDecoder(bytes.NewReader(rawYaml))
		if err := decoder.Decode(&data); err != nil {
			return fmt.Errorf("error parsing YAML: %w", err)
		}

		flatKeys := flattenYAMLKeys(data, deletePrefix)
		if len(flatKeys) == 0 {
			fmt.Println("No parameters found in the YAML file.")
			return nil
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

		fmt.Printf("The following %d parameters will be deleted from SSM:\n", len(flatKeys))

		typedKeys := make(map[string]types.ParameterType)
		for _, key := range flatKeys {
			out, err := client.GetParameter(ctx, &ssm.GetParameterInput{
				Name:           aws.String(key),
				WithDecryption: false,
			})
			if err == nil {
				typedKeys[key] = out.Parameter.Type
			}
		}

		for _, key := range flatKeys {
			lockIcon := ""
			if typedKeys[key] == types.ParameterTypeSecureString {
				lockIcon = " 🔒"
			}
			fmt.Printf("%s%s\n", color.New(color.FgHiBlack, color.Bold).Sprint(key), lockIcon)
		}

		if !deleteYes {
			fmt.Print("Are you sure? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))
			if input != "y" && input != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		for _, key := range flatKeys {
			_, err := client.DeleteParameter(ctx, &ssm.DeleteParameterInput{
				Name: aws.String(key),
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to delete %s: %v\n", color.New(color.FgWhite, color.Bold).Sprint(key), color.New(color.FgRed).Sprint(extractMessage(err)))
			} else {
				fmt.Printf("✅ Deleted %s\n", key)
			}
		}

		return nil
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&deleteFile, "file", "f", "", "Path to YAML file (required)")
	deleteCmd.Flags().StringVarP(&deletePrefix, "prefix", "p", "", "SSM prefix to delete under (required)")
	deleteCmd.Flags().BoolVarP(&deleteYes, "yes", "y", false, "Skip confirmation prompt")
}

func flattenYAMLKeys(data interface{}, prefix string) []string {
	var keys []string
	var walk func(interface{}, string)
	walk = func(node interface{}, path string) {
		switch val := node.(type) {
		case map[string]interface{}:
			for k, v := range val {
				newPath := fmt.Sprintf("%s/%s", path, k)
				walk(v, newPath)
			}
		case []interface{}:
			for i, v := range val {
				newPath := fmt.Sprintf("%s/%d", path, i)
				walk(v, newPath)
			}
		default:
			keys = append(keys, path)
		}
	}
	walk(data, strings.TrimSuffix(prefix, "/"))
	return keys
}
