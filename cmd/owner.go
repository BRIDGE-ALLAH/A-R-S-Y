package cmd

import (
	"errors"
	"fmt"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/security"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	DefaultOwnerIPsWhitelistMiddlewarePriority = DefaultRateLimitMiddlewarePriority + 5
	DefaultOwnerIPsWhitelistMiddlewareId       = "arsyOwnerIPsWhitelist"

	DefaultRequireOwnerAuthMiddlewareId             = "arsyRequireOwnerAuth"
	DefaultRequireAdminOrOwnerAuthMiddlewareId      = "arsyRequireAdminOrOwnerAuth"
)

func init() {
	// Register owner command tree
	app := core.App()
	app.OnCommand("owner", "Manage A-R-S-Y owner accounts", ownerCommand)
}

func ownerCommand(e *core.CommandEvent) error {
	if e.Args().Len() < 1 {
		return fmt.Errorf("owner command requires a subcommand")
	}

	switch e.Args().At(0) {
	case "create":
		return ownerCreateCmd(e)
	case "update":
		return ownerUpdateCmd(e)
	case "delete":
		return ownerDeleteCmd(e)
	case "list":
		return ownerListCmd(e)
	case "ips":
		return ownerIPsCommand(e)
	default:
		return fmt.Errorf("unknown owner command: %s", e.Args().At(0))
	}
}

func ownerCreateCmd(e *core.CommandEvent) error {
	if e.Args().Len() != 2 {
		return errors.New("owner create requires email and password arguments")
	}

	if e.Args().At(0) == "" || is.EmailFormat.Validate(e.Args().At(0)) != nil {
		return errors.New("missing or invalid email address")
	}

	superusersCol, err := e.App.FindCachedCollectionByNameOrId(core.CollectionNameAdmins)
	if err != nil {
		return fmt.Errorf("failed to fetch %q collection: %w", core.CollectionNameAdmins, err)
	}

	owner := core.NewRecord(superusersCol)
	owner.SetEmail(e.Args().At(0))
	owner.SetPassword(e.Args().At(1))

	if err := e.App.Save(owner); err != nil {
		return fmt.Errorf("failed to create new owner account: %w", err)
	}

	color.Green("Successfully created new owner %q!", owner.Email())
	return nil
}

func ownerUpdateCmd(e *core.CommandEvent) error {
	if e.Args().Len() != 2 {
		return errors.New("owner update requires email and password arguments")
	}

	if e.Args().At(0) == "" || is.EmailFormat.Validate(e.Args().At(0)) != nil {
		return errors.New("missing or invalid email address")
	}

	owner, err := e.App.FindAuthRecordByEmail(core.CollectionNameAdmins, e.Args().At(0))
	if err != nil {
		return fmt.Errorf("owner with email %q doesn't exist", e.Args().At(0))
	}

	owner.SetPassword(e.Args().At(1))

	if err := e.App.Save(owner); err != nil {
		return fmt.Errorf("failed to change owner %q password: %w", owner.Email(), err)
	}

	color.Green("Successfully changed owner %q password!", owner.Email())
	return nil
}

func ownerDeleteCmd(e *core.CommandEvent) error {
	if e.Args().Len() == 0 || e.Args().At(0) == "" || is.EmailFormat.Validate(e.Args().At(0)) != nil {
		return errors.New("invalid or missing email address")
	}

	owner, err := e.App.FindAuthRecordByEmail(core.CollectionNameAdmins, e.Args().At(0))
	if err != nil {
		color.Yellow("owner %q is missing or already deleted", e.Args().At(0))
		return nil
	}

	if err := e.App.Delete(owner); err != nil {
		return fmt.Errorf("failed to delete owner %q: %w", owner.Email(), err)
	}

	color.Green("Successfully deleted owner %q!", owner.Email())
	return nil
}

func ownerListCmd(e *core.CommandEvent) error {
	ownersCol, err := e.App.FindCachedCollectionByNameOrId(core.CollectionNameAdmins)
	if err != nil {
		return fmt.Errorf("failed to fetch %q collection: %w", core.CollectionNameAdmins, err)
	}

	owners, err := e.App.FindAllRecords(ownersCol)
	if err != nil {
		return fmt.Errorf("failed to list owners: %w", err)
	}

	for _, owner := range owners {
		color.Green("Owner: %s (%s)", owner.Email(), owner.Id)
	}

	return nil
}

func ownerIPsCommand(e *core.CommandEvent) error {
	settings := e.App.Settings()
	settings.OwnerIPs = e.Args().Args()
	if err := e.App.Save(settings); err != nil {
		return err
	}

	if len(e.Args().Args()) == 0 {
		color.Green("Successfully cleared OwnerIPs setting!")
	} else {
		color.New(color.BgGreen, color.FgBlack).Println("Successfully updated OwnerIPs setting:")
		ownerIPs := e.App.Settings().OwnerIPs
		for i, ip := range ownerIPs {
			if i == len(ownerIPs)-1 {
				color.Green("└─ %s", ip)
			} else {
				color.Green("├─ %s", ip)
			}
		}
	}
	return nil
}

func ownerUpsertCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:          "upsert",
		Example:      "owner upsert test@example.com 1234567890",
		Short:        "Creates, or updates if email exists, a single owner",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("missing email and password arguments")
			}

			if args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {
				return errors.New("missing or invalid email address")
			}

			ownersCol, err := app.FindCachedCollectionByNameOrId(core.CollectionNameAdmins)
			if err != nil {
				return fmt.Errorf("failed to fetch %q collection: %w", core.CollectionNameAdmins, err)
			}

			owner, err := app.FindAuthRecordByEmail(ownersCol, args[0])
			if err != nil {
				owner = core.NewRecord(ownersCol)
			}

			owner.SetEmail(args[0])
			owner.SetPassword(args[1])

			if err := app.Save(owner); err != nil {
				return fmt.Errorf("failed to upsert owner account: %w", err)
			}

			color.Green("Successfully saved owner %q!", owner.Email())
			return nil
		},
	}

	return command
}

func ownerCreateCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:          "create",
		Example:      "owner create test@example.com 1234567890",
		Short:        "Creates a new owner",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("missing email and password arguments")
			}

			if args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {

			return errors.New("missing or invalid email address")
			}

			ownersCol, err := app.FindCachedCollectionByNameOrId(core.CollectionNameAdmins)
			if err != nil {
				return fmt.Errorf("failed to fetch %q collection: %w", core.CollectionNameAdmins, err)
			}

			owner := core.NewRecord(ownersCol)
			owner.SetEmail(args[0])
			owner.SetPassword(args[1])

			if err := app.Save(owner); err != nil {
				return fmt.Errorf("failed to create new owner account: %w", err)
			}

			color.Green("Successfully created new owner %q!", owner.Email())
			return nil
		},
	}

	return command
}

func ownerUpdateCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:          "update",
		Example:      "owner update test@example.com 1234567890",
		Short:        "Changes the password of a single owner",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("missing email and password arguments")
			}

			if args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {
				return errors.New("missing or invalid email address")
			}

			owner, err := app.FindAuthRecordByEmail(core.CollectionNameAdmins, args[0])
			if err != nil {
				return fmt.Errorf("owner with email %q doesn't exist", args[0])
			}

			owner.SetPassword(args[1])

			if err := app.Save(owner); err != nil {
				return fmt.Errorf("failed to change owner %q password: %w", owner.Email(), err)
			}

			color.Green("Successfully changed owner %q password!", owner.Email())
			return nil
		},
	}

	return command
}

func ownerDeleteCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:          "delete",
		Example:      "owner delete test@example.com",
		Short:        "Deletes an existing owner",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) == 0 || args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {
				return errors.New("invalid or missing email address")
			}

			owner, err := app.FindAuthRecordByEmail(core.CollectionNameAdmins, args[0])
			if err != nil {
				color.Yellow("owner %q is missing or already deleted", args[0])
				return nil
			}

			if err := app.Delete(owner); err != nil {
				return fmt.Errorf("failed to delete owner %q: %w", owner.Email(), err)
			}

			color.Green("Successfully deleted owner %q!", owner.Email())
			return nil
		},
	}

	return command
}

func ownerListCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:   "list",
		Short: "List all owners",
		RunE:  func(command *cobra.Command, args []string) error {
			ownersCol, err := app.FindCachedCollectionByNameOrId(core.CollectionNameAdmins)
			if err != nil {
				return fmt.Errorf("failed to fetch %q collection: %w", core.CollectionNameAdmins, err)
			}

			owners, err := app.FindAllRecords(ownersCol)
			if err != nil {
				return fmt.Errorf("failed to list owners: %w", err)
			}

			for _, owner := range owners {
				color.Green("Owner: %s (%s)", owner.Email(), owner.Id)
			}

			return nil
		},
	}

	return command
}

func ownerIPsCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:          "ips",
		Example:      "owner ips 127.0.0.1 10.0.0.0/24",
		Short:        "Updates the owner IPs whitelist setting (the IPs/subnets arguments must be space separated; leave empty to clear the whitelist restriction)",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			settings := app.Settings()
			settings.OwnerIPs = args

			if err := app.Save(settings); err != nil {
				return err
			}

			if len(args) == 0 {
				color.Green("Successfully cleared OwnerIPs setting!")
			} else {
				color.New(color.BgGreen, color.FgBlack).Println("Successfully updated OwnerIPs setting:")
				ownerIPs := app.Settings().OwnerIPs
				for i, ip := range ownerIPs {
					if i == len(ownerIPs)-1 {
						color.Green("└─ %s", ip)
					} else {
						color.Green("├─ %s", ip)
					}
				}
			}
			return nil
		},
	}

	return command
}