package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	diffCmdKongStateFile   []string
	diffCmdParallelism     int
	diffCmdNonZeroExitCode bool
	diffWorkspace          string
	diffJSONOutput         bool
)

func executeDiff(cmd *cobra.Command, _ []string) error {
	return syncMain(cmd.Context(), diffCmdKongStateFile, true,
		diffCmdParallelism, 0, diffWorkspace, diffJSONOutput)
}

// newDiffCmd represents the diff command
func newDiffCmd(deprecated bool) *cobra.Command {
	use := "diff [flags] [kong-state-files...]"
	short := "Diff the current entities in Kong with the one on disks"
	execute := executeDiff
	argsValidator := cobra.MinimumNArgs(0)
	preRun := func(cmd *cobra.Command, args []string) error {
		diffCmdKongStateFile = args
		if len(diffCmdKongStateFile) == 0 {
			diffCmdKongStateFile = []string{"-"}
		}
		return preRunSilenceEventsFlag()
	}

	if deprecated {
		use = "diff"
		short = "[deprecated] use 'gateway diff' instead"
		execute = func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(os.Stderr, "Warning: 'deck diff' is DEPRECATED and will be removed in a future version. "+
				"Use 'deck gateway diff' instead.\n")
			return executeDiff(cmd, args)
		}
		argsValidator = validateNoArgs
		preRun = func(cmd *cobra.Command, args []string) error {
			if len(diffCmdKongStateFile) == 0 {
				return fmt.Errorf("a state file with Kong's configuration " +
					"must be specified using `-s`/`--state` flag")
			}
			return preRunSilenceEventsFlag()
		}
	}

	diffCmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long: `The diff command is similar to a dry run of the 'decK kong sync' command.

It loads entities from Kong and performs a diff with
the entities in local files. This allows you to see the entities
that will be created, updated, or deleted.
`,
		Args:    argsValidator,
		RunE:    execute,
		PreRunE: preRun,
	}

	if deprecated {
		diffCmd.Flags().StringSliceVarP(&diffCmdKongStateFile,
			"state", "s", []string{"kong.yaml"}, "file(s) containing Kong's configuration.\n"+
				"This flag can be specified multiple times for multiple files.\n"+
				"Use `-` to read from stdin.")
	}
	diffCmd.Flags().StringVarP(&diffWorkspace, "workspace", "w",
		"", "Diff configuration with a specific workspace "+
			"(Kong Enterprise only).\n"+
			"This takes precedence over _workspace fields in state files.")
	diffCmd.Flags().BoolVar(&dumpConfig.SkipConsumers, "skip-consumers",
		false, "do not diff consumers or "+
			"any plugins associated with consumers")
	diffCmd.Flags().IntVar(&diffCmdParallelism, "parallelism",
		10, "Maximum number of concurrent operations.")
	diffCmd.Flags().BoolVar(&noMaskValues, "no-mask-deck-env-vars-value",
		false, "do not mask DECK_ environment variable values at diff output.")
	diffCmd.Flags().StringSliceVar(&dumpConfig.SelectorTags,
		"select-tag", []string{},
		"only entities matching tags specified via this flag are diffed.\n"+
			"When this setting has multiple tag values, entities must match each of them.")
	diffCmd.Flags().BoolVar(&dumpConfig.RBACResourcesOnly, "rbac-resources-only",
		false, "sync only the RBAC resources (Kong Enterprise only).")
	diffCmd.Flags().BoolVar(&diffCmdNonZeroExitCode, "non-zero-exit-code",
		false, "return exit code 2 if there is a diff present,\n"+
			"exit code 0 if no diff is found,\n"+
			"and exit code 1 if an error occurs.")
	diffCmd.Flags().BoolVar(&dumpConfig.SkipCACerts, "skip-ca-certificates",
		false, "do not diff CA certificates.")
	diffCmd.Flags().BoolVar(&diffJSONOutput, "json-output",
		false, "generate command execution report in a JSON format")
	addSilenceEventsFlag(diffCmd.Flags())
	return diffCmd
}
