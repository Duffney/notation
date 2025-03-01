package policy

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation/cmd/notation/internal/cmdutil"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/spf13/cobra"
)

type importOpts struct {
	filePath string
	force    bool
}

func importCmd() *cobra.Command {
	var opts importOpts
	command := &cobra.Command{
		Use:   "import [flags] <file_path>",
		Short: "[Preview] Import trust policy configuration from a JSON file",
		Long: `[Preview] Import trust policy configuration from a JSON file.

** This command is in preview and under development. **

Example - Import trust policy configuration from a file:
  notation policy import my_policy.json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.filePath = args[0]
			return runImport(cmd, opts)
		},
	}
	command.Flags().BoolVar(&opts.force, "force", false, "override the existing trust policy configuration, never prompt")
	return command
}

func runImport(command *cobra.Command, opts importOpts) error {
	// optional confirmation
	if !opts.force {
		if _, err := trustpolicy.LoadDocument(); err == nil {
			confirmed, err := cmdutil.AskForConfirmation(os.Stdin, "Existing trust policy configuration found, do you want to overwrite it?", opts.force)
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "Warning: existing trust policy configuration file will be overwritten")
	}

	// read configuration
	policyJSON, err := os.ReadFile(opts.filePath)
	if err != nil {
		return fmt.Errorf("failed to read trust policy file: %w", err)
	}

	// parse and validate
	var doc trustpolicy.Document
	if err = json.Unmarshal(policyJSON, &doc); err != nil {
		return fmt.Errorf("failed to parse trust policy configuration: %w", err)
	}
	if err = doc.Validate(); err != nil {
		return fmt.Errorf("failed to validate trust policy: %w", err)
	}

	// write
	policyPath, err := dir.ConfigFS().SysPath(dir.PathTrustPolicy)
	if err != nil {
		return fmt.Errorf("failed to obtain path of trust policy file: %w", err)
	}
	if err = osutil.WriteFile(policyPath, policyJSON); err != nil {
		return fmt.Errorf("failed to write trust policy file: %w", err)
	}
	_, err = fmt.Fprintln(os.Stdout, "Trust policy configuration imported successfully.")
	return err
}
