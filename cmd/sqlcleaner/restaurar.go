package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/caiogouveia/sqlcleaner-go/internal/i18n"
	"github.com/caiogouveia/sqlcleaner-go/internal/sqldump"
	"github.com/caiogouveia/sqlcleaner-go/internal/sqlrestore"
	"github.com/spf13/cobra"
)

func newRestaurarCmd() *cobra.Command {
	var dsn string
	var createDB bool
	var dryRun bool
	var yes bool
	var jobs int

	cmd := &cobra.Command{
		Use:   "restaurar <entrada>",
		Short: i18n.T("restaurar.short"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRestaurar(os.Stdout, os.Stdin, args[0], restaurarOptions{
				dsn:      dsn,
				createDB: createDB,
				dryRun:   dryRun,
				yes:      yes,
				jobs:     jobs,
			})
		},
	}
	cmd.Flags().StringVar(&dsn, "dsn", "", i18n.T("restaurar.flag.dsn"))
	cmd.Flags().BoolVar(&createDB, "create-db", false, i18n.T("restaurar.flag.create_db"))
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, i18n.T("restaurar.flag.dry_run"))
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, i18n.T("restaurar.flag.yes"))
	cmd.Flags().IntVar(&jobs, "jobs", 4, i18n.T("restaurar.flag.jobs"))
	return cmd
}

type restaurarOptions struct {
	dsn      string
	createDB bool
	dryRun   bool
	yes      bool
	jobs     int
}

func runRestaurar(w io.Writer, in io.Reader, inputPath string, opts restaurarOptions) error {
	if _, err := os.Stat(inputPath); err != nil {
		fmt.Fprintf(w, i18n.T("err.file_not_found"), inputPath)
		return nil
	}

	fmt.Fprintln(w, "--- SQL Restaurar 1.0 ---")
	fmt.Fprintf(w, i18n.T("label.entrada"), inputPath)
	fmt.Fprintln(w, "-------------------------")

	start := time.Now()

	plainPath, cleanup, err := sqlrestore.PreparePlainFile(inputPath)
	if err != nil {
		fmt.Fprintf(w, i18n.T("err.generic"), err)
		return nil
	}
	defer cleanup()

	segments, err := sqlrestore.BuildIndex(plainPath)
	if err != nil {
		fmt.Fprintf(w, i18n.T("err.generic"), err)
		return nil
	}

	stats := sqlrestore.Summarize(segments)
	fmt.Fprintf(w, i18n.T("restaurar.summary.tables"), len(stats))
	for _, st := range stats {
		fmt.Fprintf(w, "  %-40s %10s\n", st.Table.String(), sqldump.FormatSize(float64(st.Bytes)))
	}

	cfg, err := sqlrestore.ParseConfig(opts.dsn)
	if err != nil {
		fmt.Fprintf(w, i18n.T("err.generic"), err)
		return nil
	}

	ctx := context.Background()

	if opts.createDB {
		if err := sqlrestore.CreateDatabase(ctx, cfg); err != nil {
			fmt.Fprintf(w, i18n.T("err.generic"), err)
			return nil
		}
	}

	if opts.dryRun {
		fmt.Fprintln(w, i18n.T("restaurar.dry_run.done"))
		return nil
	}

	hasData, err := sqlrestore.HasUserTables(ctx, cfg)
	if err != nil {
		fmt.Fprintf(w, i18n.T("err.generic"), err)
		return nil
	}
	if hasData && !opts.yes {
		fmt.Fprintf(w, i18n.T("restaurar.confirm.prompt"), cfg.Database)
		reader := bufio.NewReader(in)
		answer, _ := reader.ReadString('\n')
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer != "y" && answer != "yes" && answer != "s" && answer != "sim" {
			fmt.Fprintln(w, i18n.T("restaurar.confirm.aborted"))
			return nil
		}
	}

	jobs := opts.jobs
	if jobs < 1 {
		jobs = 1
	}

	prog := sqlrestore.NewPrintProgress(w)
	if err := sqlrestore.Apply(ctx, cfg, plainPath, segments, sqlrestore.Options{Jobs: jobs}, prog); err != nil {
		fmt.Fprintf(w, i18n.T("err.generic"), err)
		return nil
	}

	duration := time.Since(start).Seconds()
	fmt.Fprintf(w, i18n.T("result.done"), duration)

	return nil
}
