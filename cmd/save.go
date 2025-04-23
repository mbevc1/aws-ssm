package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	outFile    string
	savePrefix string
	rawOutput  bool
)

var saveCmd = &cobra.Command{
	Use:     "save",
	Short:   "Read parameters from AWS SSM and output YAML (use \"-\", empty, or omit to write to stdout)",
	Aliases: []string{"s", "sa"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if savePrefix == "" {
			return fmt.Errorf("--prefix is required")
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
		params, err := fetchAllParameters(savePrefix, client)
		if err != nil {
			return err
		}

		nested := flattenToNestedMap(params, savePrefix)
		return writeYAML(nested, outFile)
	},
}

func init() {
	saveCmd.Flags().StringVarP(&savePrefix, "prefix", "p", "", "SSM path prefix to read from (e.g. /myapp) (required)")
	saveCmd.Flags().StringVarP(&outFile, "out", "o", "", "Output YAML file")
	saveCmd.Flags().BoolVar(&rawOutput, "raw", false, "Disable list conversion, output all maps")
}

func fetchAllParameters(prefix string, client *ssm.Client) (map[string]string, error) {
	results := make(map[string]string)
	nextToken := aws.String("")

	for {
		input := &ssm.GetParametersByPathInput{
			Path:           aws.String(prefix),
			Recursive:      true,
			WithDecryption: true,
			NextToken:      nil,
		}
		if *nextToken != "" {
			input.NextToken = nextToken
		}

		out, err := client.GetParametersByPath(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("error fetching parameters: %w", err)
		}

		for _, param := range out.Parameters {
			results[*param.Name] = *param.Value
		}

		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}

	return results, nil
}

func flattenToNestedMap(flat map[string]string, prefix string) map[string]interface{} {
	tree := make(map[string]interface{})

	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}
	//sort.Strings(keys)
	sort.SliceStable(keys, func(i, j int) bool {
		//return len(strings.Split(keys[i], "/")) < len(strings.Split(keys[j], "/"))
		depthI := len(strings.Split(keys[i], "/"))
		depthJ := len(strings.Split(keys[j], "/"))
		if depthI != depthJ {
			return depthI < depthJ
		}
		return keys[i] < keys[j] // fallback to alphabetical within same depth
	})

	for _, fullKey := range keys {
		val := parseTypedValue(flat[fullKey])
		relativePath := strings.TrimPrefix(fullKey, prefix)
		parts := strings.Split(strings.Trim(relativePath, "/"), "/")
		insertIntoTree(tree, parts, val)
	}

	if rawOutput {
		return tree
	}
	return convertMapsToSlices(tree).(map[string]interface{})
}

func insertIntoTree(tree map[string]interface{}, parts []string, val interface{}) {
	current := tree
	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = val
			return
		}
		if _, ok := current[part]; !ok {
			current[part] = make(map[string]interface{})
		}
		current = current[part].(map[string]interface{})
	}
}

func convertMapsToSlices(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		if isNumericMap(v) {
			max := -1
			temp := make(map[int]interface{})
			for k, val := range v {
				idx, _ := strconv.Atoi(k)
				temp[idx] = convertMapsToSlices(val)
				if idx > max {
					max = idx
				}
			}
			slice := make([]interface{}, max+1)
			for i := 0; i <= max; i++ {
				slice[i] = temp[i]
			}
			return slice
		}
		for k, val := range v {
			v[k] = convertMapsToSlices(val)
		}
		return v
	case []interface{}:
		for i, val := range v {
			v[i] = convertMapsToSlices(val)
		}
		return v
	default:
		return v
	}
}

func isNumericMap(m map[string]interface{}) bool {
	if len(m) == 0 {
		return false
	}
	for k := range m {
		if _, err := strconv.Atoi(k); err != nil {
			return false
		}
	}
	return true
}

func parseTypedValue(s string) interface{} {
	s = strings.TrimSpace(s)
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

func writeYAML(data map[string]interface{}, path string) error {
	var out *os.File
	var err error

	if path == "" || outFile == "-" {
		out = os.Stdout
	} else {
		out, err = os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer out.Close()
	}

	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)
	return enc.Encode(data)
}
