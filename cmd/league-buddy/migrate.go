package main

import (
	"context"
	"fmt"
	"os"

	"github.com/HatiCode/league-buddy/internal/store"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run pending database migrations or roll back the last migration`,
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := getDB()
		if err != nil {
			return err
		}
		defer db.Close()

		if err := store.Migrate(db.DB()); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		fmt.Println("Migrations completed successfully")
		return nil
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Roll back the last migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := getDB()
		if err != nil {
			return err
		}
		defer db.Close()

		if err := store.MigrateDown(db.DB()); err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}

		fmt.Println("Rollback completed successfully")
		return nil
	},
}

func getDB() (*store.PostgresStore, error) {
	url := dbURL
	if url == "" {
		url = os.Getenv("DATABASE_URL")
	}
	if url == "" {
		return nil, fmt.Errorf("database URL required: use --db-url or set DATABASE_URL")
	}

	return store.NewPostgresStore(context.Background(), url)
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	rootCmd.AddCommand(migrateCmd)
}
