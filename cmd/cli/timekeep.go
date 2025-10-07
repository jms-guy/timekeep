package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (s *CLIService) addProgramsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "add",
		Aliases: []string{"Add", "ADD"},
		Short:   "Add a program to begin tracking",
		Long:    "User may specify any number of programs to track in a single command, as long as they're separated by a space",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			return s.AddPrograms(ctx, args)
		},
	}
}

func (s *CLIService) removeProgramsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm",
		Aliases: []string{"RM", "remove", "Remove", "REMOVE"},
		Short:   "Remove a program from tracking list",
		Long:    "User may specify multiple programs to remove, as long as they're separated by a space. May provide the --all flag to remove all programs from tracking list",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			all, _ := cmd.Flags().GetBool("all")

			return s.RemovePrograms(ctx, args, all)
		},
	}

	cmd.Flags().Bool("all", false, "Removes all currently tracked programs")

	return cmd
}

func (s *CLIService) getListcmd() *cobra.Command {
	return &cobra.Command{
		Use:     "ls",
		Aliases: []string{"LS", "list", "List", "LIST"},
		Short:   "Lists programs being tracked by service",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			return s.GetList(ctx)
		},
	}
}

func (s *CLIService) infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "info",
		Aliases: []string{"Info", "INFO"},
		Short:   "Shows basic info for currently tracked programs",
		Long:    "Accepts program name as an argument to show in depth stats for that program, else shows basic stats for all programs",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if len(args) == 0 {
				return s.GetAllInfo(ctx)
			} else {
				return s.GetInfo(ctx, args)
			}
		},
	}
}

func (s *CLIService) sessionHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "history",
		Aliases: []string{"History", "HISTORY"},
		Short:   "Shows session history",
		Long:    "If no args given, shows previous 25 sessions. Program name may be given as argument to filter only those sessions. Flags may be given to filter further, with OR without program name",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			date, _ := cmd.Flags().GetString("date")
			start, _ := cmd.Flags().GetString("start")
			end, _ := cmd.Flags().GetString("end")
			limit, _ := cmd.Flags().GetInt64("limit")

			return s.GetSessionHistory(ctx, args, date, start, end, limit)
		},
	}

	cmd.Flags().String("date", "", "Filter session history by date")
	cmd.Flags().String("start", "", "Filters session history by adding a starting date")
	cmd.Flags().String("end", "", "Filters session history by adding an ending date")
	cmd.Flags().Int64("limit", 25, "Adjusts number limit of sessions shown")

	return cmd
}

func (s *CLIService) refreshCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "refresh",
		Aliases: []string{"Refresh", "REFRESH"},
		Short:   "Sends a manual refresh command to the service",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := s.ServiceCmd.WriteToService()
			if err != nil {
				return err
			}
			fmt.Println("Service refresh command sent successfully")
			return nil
		},
	}
}

func (s *CLIService) resetStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "reset",
		Aliases: []string{"Reset", "RESET"},
		Short:   "Reset tracking stats",
		Long:    "Reset tracking stats for given programs, accepts multiple programs with a space between them. May provide the --all flag to reset all stats",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			all, _ := cmd.Flags().GetBool("all")

			return s.ResetStats(ctx, args, all)
		},
	}

	cmd.Flags().Bool("all", false, "Resets all currently tracked program data. Does not remove programs from tracking")

	return cmd
}

func (s *CLIService) statusServiceCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Aliases: []string{"Status", "STATUS"},
		Short:   "Gets current OS state of Timekeep service",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.StatusService()
		},
	}
}

func (s *CLIService) getActiveSessionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "active",
		Aliases: []string{"Active", "ACTIVE"},
		Short:   "Get list of current active sessions being tracked",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			return s.GetActiveSessions(ctx)
		},
	}
}

func (s *CLIService) getVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"Version", "VERSION"},
		Short:   "Get current version of Timekeep",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.GetVersion()
		},
	}
}

func (s *CLIService) wakatimeIntegration() *cobra.Command {
	return &cobra.Command{
		Use:     "wakatime",
		Aliases: []string{"WakaTime", "WAKATIME"},
		Short:   "Enable/disable integration with WakaTime",
	}
}

func (s *CLIService) wakatimeEnable() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "enable",
		Aliases: []string{"Enable", "ENABLE"},
		Short:   "Enable WakaTime integration",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey, _ := cmd.Flags().GetString("api-key")

			return s.EnableWakaTime(apiKey)
		},
	}

	cmd.Flags().String("api-key", "", "User's WakaTime API key")

	return cmd
}

func (s *CLIService) wakatimeDisable() *cobra.Command {
	return &cobra.Command{
		Use:     "disable",
		Aliases: []string{"Disable", "DISABLE"},
		Short:   "Disable WakaTime integration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.DisableWakaTime()
		},
	}
}
