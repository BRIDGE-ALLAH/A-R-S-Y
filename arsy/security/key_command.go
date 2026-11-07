package security

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// NewGenerateKeyCommand creates an explicit, one-time secret-key display command.
func NewGenerateKeyCommand(stdout io.Writer) *cobra.Command {
	if stdout == nil {
		stdout = io.Discard
	}
	return &cobra.Command{
		Use:   "generate-secret-key",
		Short: "Generate and display a new A-R-S-Y encryption key",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := GenerateKey()
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(stdout, key); err != nil {
				return fmt.Errorf("write generated key: %w", err)
			}
			return nil
		},
	}
}
