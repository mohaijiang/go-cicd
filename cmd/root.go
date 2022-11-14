package cmd

import (
	"fmt"
	"os"
	"state-example/pkg/engine"
	"state-example/pkg/logger"
	"state-example/pkg/stream"

	"github.com/spf13/cobra"
)

var (
	yamlFile string
)

var rootCmd = &cobra.Command{
	Use: "state-example",
	Run: func(_ *cobra.Command, _ []string) {

		go stream.Output()

		job := engine.GetJob(yamlFile)
		for name := range job.Stages {
			logger.Info(name)
		}

		engine.Engine(job)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&yamlFile, "file", "f", "cicd.yaml", "yaml config file")
}
