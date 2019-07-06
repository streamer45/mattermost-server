// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattermost/mattermost-server/store"
	"github.com/spf13/cobra"
)

var IntegrityCmd = &cobra.Command{
	Use:   "integrity",
	Short: "Check database data integrity",
	RunE:  integrityCmdF,
}

func init() {
	IntegrityCmd.Flags().Bool("confirm", false, "Confirm you really want to run a complete integrity check of your DB.")
	IntegrityCmd.Flags().BoolP("verbose", "v", false, "Show detailed information on integrity check results")
	RootCmd.AddCommand(IntegrityCmd)
}

func printIntegrityCheckResult(result store.IntegrityCheckResult, verbose bool) {
	fmt.Println(fmt.Sprintf("Found %d records in relation %s orphans of relation %s",
		len(result.Records), result.Info.ChildName, result.Info.ParentName))
	if !verbose {
		return
	}
	for _, record := range result.Records {
		fmt.Println(fmt.Sprintf("  Child %s is missing Parent %s", record.ChildId, record.ParentId))
	}
}

func integrityCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	confirmFlag, _ := command.Flags().GetBool("confirm")
	if !confirmFlag {
		var confirm string
		fmt.Fprintf(os.Stdout, "This check can harm performance on live systems. Are you sure you want to proceed? (y/N): ")
		fmt.Scanln(&confirm)
		if !strings.EqualFold(confirm, "y") && !strings.EqualFold(confirm, "yes") {
			fmt.Fprintf(os.Stderr, "Aborted.\n")
			return nil
		}
	}

	verboseFlag, _ := command.Flags().GetBool("verbose")
	results := a.Srv.Store.CheckIntegrity()
	for result := range results {
		if result.Err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", result.Err.Error())
			break
		}
		printIntegrityCheckResult(result, verboseFlag)
	}
	return nil
}