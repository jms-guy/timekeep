package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (s *CLIService) addProgramsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "add",
		Aliases: []string{"Add", "ADD"},
		Short:   "Add a program name to begin tracking",
		Long:    "User may specify any number of programs to track in a single command, as long as they're seperated by a space",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.AddPrograms(args)
		},
	}
}

func (s *CLIService) removeProgramsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove",
		Aliases: []string{"Remove", "REMOVE"},
		Short:   "Remove a program from tracking list",
		Long:    "User may specify multiple programs to remove, as long as they're seperated by a space. May provide the --all flag to remove all programs from tracking list",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")

			return s.RemovePrograms(args, all)
		},
	}

	cmd.Flags().Bool("all", false, "When provided removes all currently tracked programs")

	return cmd
}

func (s *CLIService) getListcmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"List", "LIST"},
		Short:   "Lists programs being tracked by service",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.GetList()
		},
	}
}

func (s *CLIService) statsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "stats",
		Aliases: []string{"Stats", "STATS"},
		Short:   "Shows stats for currently tracked programs",
		Long:    "Accepts program name as an argument to show in depth stats for that program, else shows basic stats for all programs",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return s.GetAllStats()
			} else {
				return s.GetStats(args)
			}
		},
	}
}

func (s *CLIService) sessionHistoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "history",
		Aliases: []string{"History", "HISTORY"},
		Short:   "Shows session history for a given program name",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.GetSessionHistory(args)
		},
	}
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
			all, _ := cmd.Flags().GetBool("all")

			return s.ResetStats(args, all)
		},
	}

	cmd.Flags().Bool("all", false, "If flag is provided, resets all currently tracked program data. Does not remove programs from tracking")

	return cmd
}

func (s *CLIService) pingServiceCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "ping",
		Aliases: []string{"Query", "QUERY"},
		Short:   "Gets current OS state of Timekeep service",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.PingService()
		},
	}
}
