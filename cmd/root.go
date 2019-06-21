package cmd

import (
	"fmt"
	"os"

	"github.com/johnnymn/risr/cmd/deploy"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "risr <command>",
	Short: "Zero downtime deploy script for Amazon Machine Images (AMI)",
	Long: `
riser allows you to deploy an AMI to a fleet of EC2 servers without
any downtime, it accomplishes this  by using a blue-green strategy
to rollout a new group of servers before dropping the old ones.
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(deploy.DeployCommand)
}
