package security

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewSecretKeyCommand creates the explicit one-time key-generation command.
func NewSecretKeyCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "secret-key",
		Short: "manage the external A-R-S-Y encryption key",
	}
	root.AddCommand(&cobra.Command{
		Use:   "generate",
		Short: "generate a random 32-byte A-R-S-Y secret key",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := GenerateKey()
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), key)
			return err
		},
	})
	return root
}
