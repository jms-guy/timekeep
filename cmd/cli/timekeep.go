package main

import "github.com/spf13/cobra"

func (s *CLIService) addProgramsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "add",
		Aliases: []string{"Add", "ADD"},
		Short:   "Add a program name to begin tracking",
		Long:    "User may specify any number of programs to track in a single command, as long as they're seperated by a space",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.addPrograms(args)
		},
	}
}

func (s *CLIService) removeProgramsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove",
		Aliases: []string{"Remove", "REMOVE"},
		Short:   "Remove a program from tracking list",
		Long:    "User may specify multiple programs to remove, as long as they're seperated by a space",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.removePrograms(args)
		},
	}
}

func (s *CLIService) getListcmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"List", "LIST"},
		Short:   "Lists programs being tracked by service",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.getList()
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
				return s.getAllStats()
			} else {
				return s.getStats(args)
			}
		},
	}
}
