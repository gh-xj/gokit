package main

import (
	"fmt"
	"os"
	"path/filepath"

	agentops "github.com/gh-xj/agentops"
	"github.com/gh-xj/agentops/resource"
	"github.com/spf13/cobra"
)

func newDoctorCmd(reg *resource.Registry, ctx *agentops.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Validate strategy and resource health",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			if dir == "" {
				dir = "."
			}

			report := agentops.DoctorReport{
				SchemaVersion: "1.0",
				OK:            true,
			}

			// Check that .agentops/ exists with required files.
			requiredFiles := []string{"storage.yaml", "transitions.yaml"}
			agentopsDir := filepath.Join(dir, ".agentops")
			for _, name := range requiredFiles {
				p := filepath.Join(agentopsDir, name)
				if _, err := os.Stat(p); err != nil {
					report.OK = false
					report.Findings = append(report.Findings, agentops.DoctorFinding{
						Code:    "missing_file",
						Path:    p,
						Message: fmt.Sprintf("required file missing: %s", name),
					})
				}
			}

			// Iterate resources and call Validate on those that support it.
			for _, res := range reg.All() {
				v, ok := res.(resource.Validator)
				if !ok {
					continue
				}
				schema := res.Schema()
				records, err := res.List(ctx, nil)
				if err != nil {
					report.Findings = append(report.Findings, agentops.DoctorFinding{
						Code:    "list_error",
						Path:    schema.Kind,
						Message: fmt.Sprintf("failed to list %s resources: %s", schema.Kind, err),
					})
					report.OK = false
					continue
				}
				for _, rec := range records {
					sub, err := v.Validate(ctx, rec.ID)
					if err != nil {
						report.Findings = append(report.Findings, agentops.DoctorFinding{
							Code:    "validate_error",
							Path:    rec.ID,
							Message: fmt.Sprintf("validate %s %s: %s", schema.Kind, rec.ID, err),
						})
						report.OK = false
						continue
					}
					if !sub.OK {
						report.OK = false
						report.Findings = append(report.Findings, sub.Findings...)
					}
				}
			}

			jsonFlag, _ := cmd.Flags().GetString("json")
			if jsonFlag != "" {
				out, err := report.JSON()
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), out)
			} else {
				if report.OK {
					fmt.Fprintln(cmd.OutOrStdout(), "doctor: ok")
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "doctor: issues found")
					for _, f := range report.Findings {
						fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s: %s\n", f.Code, f.Path, f.Message)
					}
				}
			}

			if !report.OK {
				return fmt.Errorf("doctor check failed")
			}
			return nil
		},
	}
	return cmd
}
