package main

import "github.com/spf13/cobra"

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resources from Riot API",
}

func init() {
	rootCmd.AddCommand(getCmd)
}
