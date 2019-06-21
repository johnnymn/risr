package deploy

import (
	"fmt"
	"os"

	"github.com/johnnymn/risr/pkg/deployments/v1alpha1"
	"github.com/spf13/cobra"
)

var DeployCommand = &cobra.Command{
	Use:   "deploy <filename>",
	Short: "Deploys a Stack",
	Long: `
Takes a Stack definition (in the form of a .yaml file),
and deploys it to AWS. It uses the default AWS ENV
variables for authentication to the AWS API.
`,
	Example: ` risr deploy stack.yaml`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileData, err := readStackFile(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error reading Stack file: "+err.Error())
			os.Exit(1)
		}

		stack, err := parseStackFromBytes(fileData)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error decoding Stack definition string: "+err.Error())
			os.Exit(1)
		}

		manager, err := v1alpha1.NewManager()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error instantiating deployment manager: "+err.Error())
			os.Exit(1)
		}

		err = manager.DeployStack(stack)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error deploying Stack: "+err.Error())
			os.Exit(1)
		}

		fmt.Print("Stack " + stack.Name + " was deployed successfully!!!")
	},
}
