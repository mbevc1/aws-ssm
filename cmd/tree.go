package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	decryptValues bool
	showValues    bool
	treePrefix    string
)

var treeCmd = &cobra.Command{
	Use:     "tree",
	Short:   "Print a tree structure of parameters under a given prefix",
	Aliases: []string{"t", "tr"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if treePrefix == "" {
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
		paramData, err := fetchAllParameterObjects(treePrefix, client)
		if err != nil {
			return err
		}

		paths := make([]string, 0, len(paramData))
		for k := range paramData {
			rel := strings.TrimPrefix(k, treePrefix)
			rel = strings.Trim(rel, "/")
			if rel != "" {
				paths = append(paths, rel)
			}
		}

		sort.Strings(paths)
		printTree(paths, paramData, treePrefix)
		return nil
	},
}

func init() {
	treeCmd.Flags().BoolVarP(&decryptValues, "decrypt", "d", false, "Decrypt SecureString values (requires IAM permission)")
	treeCmd.Flags().StringVarP(&treePrefix, "prefix", "p", "", "SSM path prefix to read from (e.g. /myapp) (required)")
	treeCmd.Flags().BoolVarP(&showValues, "values", "v", false, "Show values alongside keys")
}

type treeParam struct {
	Type  types.ParameterType
	Value string
}

func fetchAllParameterObjects(prefix string, client *ssm.Client) (map[string]treeParam, error) {
	result := make(map[string]treeParam)
	nextToken := aws.String("")

	for {
		input := &ssm.GetParametersByPathInput{
			Path:           aws.String(prefix),
			Recursive:      true,
			WithDecryption: decryptValues,
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
			result[*param.Name] = treeParam{
				Type:  param.Type,
				Value: *param.Value,
			}
		}

		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}

	return result, nil
}

func printTree(paths []string, values map[string]treeParam, rootPrefix string) {
	type node struct {
		name     string
		fullPath string
		children map[string]*node
	}

	root := &node{name: "/", fullPath: rootPrefix, children: make(map[string]*node)}

	for _, path := range paths {
		parts := strings.Split(path, "/")
		current := root
		currentPath := strings.TrimSuffix(rootPrefix, "/")

		for _, part := range parts {
			currentPath += "/" + part
			if current.children[part] == nil {
				current.children[part] = &node{
					name:     part,
					fullPath: currentPath,
					children: make(map[string]*node),
				}
			}
			current = current.children[part]
		}
	}

	var walk func(n *node, prefix string, last bool)
	walk = func(n *node, prefix string, last bool) {
		connector := "├── "
		if last {
			connector = "└── "
		}

		label := n.name
		// Color numbers differently (e.g., list indices)
		if _, err := strconv.Atoi(n.name); err == nil {
			label = color.New(color.FgYellow).Sprint(label)
		}

		// Apply secure string or standard coloring
		if param, ok := values[n.fullPath]; ok {
			// Format base label with SecureString icon if needed
			if param.Type == types.ParameterTypeSecureString {
				label = color.New(color.FgCyan).Sprintf("%s 🔒", label)
			} else {
				label = color.New(color.FgWhite).Sprint(label)
			}

			// Append value if requested
			if showValues {
				label += fmt.Sprintf(" = %s", color.New(color.FgHiBlack).Sprint(param.Value))
			}
		}

		if n.name != "/" {
			fmt.Printf("%s%s%s\n", prefix, connector, label)
		}

		keys := make([]string, 0, len(n.children))
		for k := range n.children {
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

		newPrefix := prefix
		if n.name != "/" {
			newPrefix += map[bool]string{true: "    ", false: "│   "}[last]
		}

		for i, k := range keys {
			walk(n.children[k], newPrefix, i == len(keys)-1)
		}
	}

	walk(root, "", true)
}
