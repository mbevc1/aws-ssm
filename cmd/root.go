package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

// Name & Version of the app
var (
	Name    = "aws-ssm"
	Version string
)

var (
	debugFlag bool
	ctx       context.Context
	awsRegion string
)

var rootCmd = &cobra.Command{
	Use:           fmt.Sprintf("%s", Name),
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	Short:         fmt.Sprintf("%s is a CLI tool for managing AWS SSM Params", Name),
	Long:          fmt.Sprintf("%s is a CLI utility for managing YAML ↔ AWS SSM Parameter Store", Name),
}

func Execute(version string) {
	Version = version
	// fmt.Println()
	// defer fmt.Println()

	rootCmd.Version = version

	cobra.CheckErr(rootCmd.Execute())

	if debugFlag {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	slog.Debug(fmt.Sprintf("App version: %s", Version))
}

func init() {
	//cobra.OnInitialize(initFunc)

	// start with empty Context
	ctx = context.Background()

	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(saveCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(treeCmd)
	rootCmd.AddCommand(yamlTreeCmd)
	//rootCmd.AddCommand(versionCmd)

	rootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "b", false, "Enable debugging logging")
	rootCmd.PersistentFlags().StringVarP(&awsRegion, "region", "r", "", "AWS region to use (overrides default profile)")
}
