package main

import (
	"fmt"
	"os"

	"github.com/AkhileshThykkat/ez-mig/internal/runner"
	"github.com/AkhileshThykkat/ez-mig/internal/session"
	"github.com/AkhileshThykkat/ez-mig/internal/store"
	"github.com/spf13/cobra"
)

var (
	dbTarget  string
	dbURIFlag string
	pathFlag  string
)
var rootCmd = &cobra.Command{
	Use:   "ez-mig",
	Short: "ez-mig is a CLI tool wrapping golang-migrate with session management",
	Long:  `A self-hostable companion tool to quickly execute database migrations using saved SQLite sessions.`,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage saved database connection sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var configAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new session configuration target",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if dbURIFlag == "" || pathFlag == "" {
			return fmt.Errorf("--uri and --path flags are required when adding a session")
		}
		s := session.Session{
			Name:               args[0],
			DbURI:              dbURIFlag,
			MigrationFilesPath: pathFlag,
		}
		if err := store.Create(s); err != nil {
			return err
		}
		fmt.Printf("Session %q successfully saved.\n", s.Name)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved config targets",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := store.List()
		if err != nil {
			return err
		}
		if len(sessions) == 0 {
			fmt.Println("No sessions saved yet. Add one using 'ez-mig config add <name> --uri <uri> --path <path>'")
			return nil
		}
		fmt.Printf("%-15s %-40s %-s\n", "NAME", "DATABASE URI", "MIGRATIONS PATH")
		fmt.Println("-------------------------------------------------------------------------------------")
		for _, s := range sessions {
			fmt.Printf("%-15s %-40s %-s\n", s.Name, s.DbURI, s.MigrationFilesPath)
		}
		return nil
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Delete a session configuration by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := store.Delete(args[0]); err != nil {
			return err
		}
		fmt.Printf("Session %q removed.\n", args[0])
		return nil
	},
}

var configUpdateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Update an existing session configuration target",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if dbURIFlag == "" && pathFlag == "" {
			return fmt.Errorf("provide at least one flag to update: --uri or --path")
		}

		existing, err := store.GetByName(args[0])
		if err != nil {
			return err
		}

		if dbURIFlag != "" {
			existing.DbURI = dbURIFlag
		}
		if pathFlag != "" {
			existing.MigrationFilesPath = pathFlag
		}

		if err := store.Update(*existing); err != nil {
			return err
		}
		fmt.Printf("Session %q successfully updated.\n", args[0])
		return nil
	},
}

func executeMigration(action runner.Action, arg *int) error {
	s, err := store.GetByName(dbTarget)
	if err != nil {
		return err
	}
	return runner.Execute(*s, action, arg)
}

var upCmd = &cobra.Command{
	Use:   "up [count]",
	Short: "Apply all or N pending migrations",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var step *int
		if len(args) == 1 {
			var val int
			if _, err := fmt.Sscanf(args[0], "%d", &val); err != nil {
				return fmt.Errorf("argument must be a valid integer step count")
			}
			step = &val
		}
		return executeMigration(runner.ActionUp, step)
	},
}

var downCmd = &cobra.Command{
	Use:   "down [count]",
	Short: "Roll back 1 or N migrations",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var step *int
		if len(args) == 1 {
			var val int
			if _, err := fmt.Sscanf(args[0], "%d", &val); err != nil {
				return fmt.Errorf("argument must be a valid integer step count")
			}
			step = &val
		}
		return executeMigration(runner.ActionDown, step)
	},
}

var gotoCmd = &cobra.Command{
	Use:   "goto <version>",
	Short: "Migrate directly to a specific version number",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var version int
		if _, err := fmt.Sscanf(args[0], "%d", &version); err != nil {
			return fmt.Errorf("version argument must be a valid integer")
		}
		return executeMigration(runner.ActionGoto, &version)
	},
}

var forceCmd = &cobra.Command{
	Use:   "force <version>",
	Short: "Force-set a specific schema version to fix a dirty state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var version int
		if _, err := fmt.Sscanf(args[0], "%d", &version); err != nil {
			return fmt.Errorf("version argument must be a valid integer")
		}
		return executeMigration(runner.ActionForce, &version)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current applied database schema version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeMigration(runner.ActionVersion, nil)
	},
}

func init() {
	cobra.OnInitialize(func() {
		if err := store.StoreSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "Initialization error: %v\n", err)
			os.Exit(1)
		}
	})

	configAddCmd.Flags().StringVar(&dbURIFlag, "uri", "", "Database raw connection URI link details")
	configAddCmd.Flags().StringVar(&pathFlag, "path", "", "Local storage operating path tracking raw scripts")

	configUpdateCmd.Flags().StringVar(&dbURIFlag, "uri", "", "Database raw connection URI link details")
	configUpdateCmd.Flags().StringVar(&pathFlag, "path", "", "Local storage operating path tracking raw scripts")

	configCmd.AddCommand(configAddCmd, configListCmd, configRemoveCmd, configUpdateCmd)

	for _, cmd := range []*cobra.Command{upCmd, downCmd, gotoCmd, forceCmd, versionCmd} {
		cmd.Flags().StringVar(&dbTarget, "db", "", "Database target session name profile (Required)")
		cmd.MarkFlagRequired("db")
		rootCmd.AddCommand(cmd)
	}

	rootCmd.AddCommand(configCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
