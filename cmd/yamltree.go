package cmd

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var yamlTreeCmd = &cobra.Command{
	Use:     "yaml-tree",
	Short:   "Print a tree structure of a YAML config file",
	Aliases: []string{"yt", "ytr"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if yamlFile == "" {
			return fmt.Errorf("--file is required")
		}

		rawYaml, err := os.ReadFile(yamlFile)
		if err != nil {
			return fmt.Errorf("error reading YAML file: %w", err)
		}

		var data map[string]interface{}
		decoder := yaml.NewDecoder(bytes.NewReader(rawYaml))
		if err := decoder.Decode(&data); err != nil {
			return fmt.Errorf("error parsing YAML: %w", err)
		}

		printYAMLTree(data)
		return nil
	},
}

func init() {
	yamlTreeCmd.Flags().StringVarP(&yamlFile, "file", "f", "", "YAML file to inspect (required)")
	yamlTreeCmd.Flags().BoolVarP(&showValues, "values", "v", false, "Show values alongside keys")
}

func printYAMLTree(data interface{}) {
	isSensitiveKey := func(key string) bool {
		keywords := []string{"password", "secret", "token", "key", "apikey", "auth", "private"}
		key = strings.ToLower(key)
		for _, word := range keywords {
			if strings.Contains(key, word) {
				return true
			}
		}
		return false
	}

	isLeaf := func(v interface{}) bool {
		switch v.(type) {
		case string, bool, int, int64, float64, float32, nil:
			return true
		default:
			return false
		}
	}

	var walk func(node interface{}, prefix string, last bool, fullPath string)
	walk = func(node interface{}, prefix string, last bool, fullPath string) {
		connector := "├── "
		if last {
			connector = "└── "
		}

		switch v := node.(type) {
		case map[string]interface{}:
			keys := make([]string, 0, len(v))
			for k := range v {
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

			for i, k := range keys {
				isLast := i == len(keys)-1
				newPrefix := prefix + map[bool]string{true: "    ", false: "│   "}[!isLast]
				nextPath := fullPath + "/" + k
				label := color.New(color.FgWhite).Sprint(k)

				lock := ""
				if isSensitiveKey(nextPath) {
					lock = " 🔒"
					label = color.New(color.FgCyan).Sprint(k)
				}
				if isLeaf(v[k]) && showValues {
					value := fmt.Sprintf("%v", v[k])
					fmt.Printf("%s%s%s%s = %s\n", prefix, connector, label, lock, color.New(color.FgHiBlack).Sprint(value))
				} else {
					fmt.Printf("%s%s%s%s\n", prefix, connector, label, lock)
					walk(v[k], newPrefix, isLast, nextPath)
				}
			}

		case []interface{}:
			for i, item := range v {
				isLast := i == len(v)-1
				newPrefix := prefix + map[bool]string{true: "    ", false: "│   "}[!isLast]
				nextPath := fmt.Sprintf("%s/%d", fullPath, i)
				label := color.New(color.FgYellow).Sprintf("%d", i)

				lock := ""
				if isSensitiveKey(nextPath) {
					lock = " 🔒"
					label = color.New(color.FgCyan).Sprint(i)
				}
				if isLeaf(item) && showValues {
					value := fmt.Sprintf("%v", item)
					fmt.Printf("%s%s%s%s = %s\n", prefix, connector, label, lock, color.New(color.FgHiBlack).Sprint(value))
				} else {
					fmt.Printf("%s%s%s%s\n", prefix, connector, label, lock)
					walk(item, newPrefix, isLast, nextPath)
				}
			}
		}
	}

	//fmt.Println(color.New(color.FgGreen, color.Bold).Sprint("└──"))
	//walk(data, "    ", true, "")
	fmt.Println(color.New(color.FgHiCyan, color.Bold).Sprint("root"))
	walk(data, "", true, "")
}
