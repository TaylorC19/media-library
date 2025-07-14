package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file - this will fail if .env doesn't exist
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// TLS設定
	var tlsOpt string
	tlsEnabled, err := strconv.ParseBool(os.Getenv("DB_TLS_ENABLED"))
	if err != nil {
		tlsEnabled = false
	}
	tlsCaFile := os.Getenv("DB_TLS_CA_FILE")
	if tlsEnabled && tlsCaFile != "" {
		tlsOpt = fmt.Sprintf("tls=true&tlsCAFile=%s", tlsCaFile)
	} else {
		tlsOpt = "tls=false"
	}

	mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin&replicaSet=rs0&readPreference=secondaryPreferred&retryWrites=false&%s",
		url.QueryEscape(os.Getenv("DB_ADMIN_USER")),
		url.QueryEscape(os.Getenv("DB_ADMIN_PASSWORD")),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		tlsOpt,
	)

	command := os.Args[1]

	

	// Check required environment variables
	requiredEnvVars := []string{"DB_MIGRATION_PATH", "DB_ADMIN_USER", "DB_ADMIN_PASSWORD", "DB_HOST", "DB_PORT", "DB_NAME"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("Required environment variable %s is not set in .env file", envVar)
		}
	}

	// Create migrate instance
	migrationsPath := os.Getenv("DB_MIGRATION_PATH")
	m, err := migrate.New("file://"+migrationsPath, mongoURI)
	if err != nil {
		log.Fatalf("Could not create migrate instance: %v", err)
	}
	defer m.Close()

	switch command {
	case "up":
		log.Println("Applying all up migrations...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		if err == migrate.ErrNoChange {
			log.Println("No new migrations to apply")
		} else {
			log.Println("Migrations applied successfully!")
		}

	case "down":
		log.Println("Rolling back one migration...")
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
		if err == migrate.ErrNoChange {
			log.Println("No migrations to rollback")
		} else {
			log.Println("Migration rolled back successfully!")
		}

	case "goto":
		if len(os.Args) < 3 {
			log.Fatal("Usage: go run ./cmd/migrate/migrate.go goto VERSION")
		}
		versionStr := os.Args[2]
		version, err := strconv.ParseUint(versionStr, 10, 64)
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		log.Printf("Migrating to version %d...\n", version)
		if err := m.Migrate(uint(version)); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to migrate to version %d: %v", version, err)
		}
		log.Println("Migration to specific version successful!")

	case "force":
		if len(os.Args) < 3 {
			log.Fatal("Usage: go run ./cmd/migrate/migrate.go force VERSION")
		}
		versionStr := os.Args[2]
		version, err := strconv.ParseInt(versionStr, 10, 64)
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		log.Printf("Forcing migration to version %d...\n", version)
		if err := m.Force(int(version)); err != nil {
			log.Fatalf("Failed to force migration to version %d: %v", version, err)
		}
		log.Println("Migration forced to specific version successfully!")

	case "create":
		if len(os.Args) < 3 {
			log.Fatal("Usage: go run ./cmd/migrate/migrate.go create <migration_name> [--format=js|json]")
		}
		// Parse format flag for create command
		formatFlag := flag.NewFlagSet("create", flag.ExitOnError)
		format := formatFlag.String("format", "js", "Migration format: js or json")
		formatFlag.Parse(os.Args[3:])
		createMigration(os.Args[2], *format)

	case "status":
		checkMigrationStatus(m)

	case "version":
		getMigrationVersion(m)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func createMigration(name, format string) {
	migrationsPath := os.Getenv("DB_MIGRATION_PATH")

	// Ensure migrations directory exists
	if err := os.MkdirAll(migrationsPath, 0755); err != nil {
		log.Fatalf("Failed to create migrations directory: %v", err)
	}

	// Create timestamp for migration filename
	timestamp := time.Now().Format("20060102150405")
	filename := timestamp + "_" + name

	switch format {
	case "js":
		createJavaScriptMigration(filename, name, migrationsPath)
	case "json":
		createJSONMigration(filename, name, migrationsPath)
	default:
		log.Fatalf("Unsupported format: %s. Use 'js' or 'json'", format)
	}
}

func createJavaScriptMigration(filename, name, migrationsPath string) {
	// Create up migration file (JavaScript)
	upFile := fmt.Sprintf("%s/%s.up.js", migrationsPath, filename)
	upContent := fmt.Sprintf(`// Up migration for %s
// Add your MongoDB operations here
// Example:
// db.createCollection("collection_name");
// db.collection_name.createIndex({ "field": 1 });
// db.collection_name.insertOne({ "field": "value" });

`, name)
	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		log.Fatalf("Failed to create up migration file: %v", err)
	}

	// Create down migration file (JavaScript)
	downFile := fmt.Sprintf("%s/%s.down.js", migrationsPath, filename)
	downContent := fmt.Sprintf(`// Down migration for %s
// Add your rollback operations here
// Example:
// db.collection_name.drop();
// db.collection_name.deleteOne({ "field": "value" });

`, name)
	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		log.Fatalf("Failed to create down migration file: %v", err)
	}

	log.Printf("Created JavaScript migration files:\n  %s\n  %s", upFile, downFile)
}

func createJSONMigration(filename, name, migrationsPath string) {
	// Create up migration file (JSON)
	upFile := fmt.Sprintf("%s/%s.up.json", migrationsPath, filename)
	upContent := map[string]interface{}{
		"operations": []map[string]interface{}{
			{
				"comment": fmt.Sprintf("Up migration for %s", name),
			},
			{
				"createCollection": "example_collection",
			},
			{
				"createIndex": map[string]interface{}{
					"collection": "example_collection",
					"index":      map[string]interface{}{"field": 1},
					"options":    map[string]interface{}{"unique": true},
				},
			},
			{
				"insertOne": map[string]interface{}{
					"collection": "example_collection",
					"document": map[string]interface{}{
						"field":      "value",
						"created_at": "2024-01-01T00:00:00Z",
					},
				},
			},
		},
	}

	upJSON, err := json.MarshalIndent(upContent, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal up migration JSON: %v", err)
	}

	if err := os.WriteFile(upFile, upJSON, 0644); err != nil {
		log.Fatalf("Failed to create up migration file: %v", err)
	}

	// Create down migration file (JSON)
	downFile := fmt.Sprintf("%s/%s.down.json", migrationsPath, filename)
	downContent := map[string]interface{}{
		"operations": []map[string]interface{}{
			{
				"comment": fmt.Sprintf("Down migration for %s", name),
			},
			{
				"deleteOne": map[string]interface{}{
					"collection": "example_collection",
					"filter":     map[string]interface{}{"field": "value"},
				},
			},
			{
				"dropCollection": "example_collection",
			},
		},
	}

	downJSON, err := json.MarshalIndent(downContent, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal down migration JSON: %v", err)
	}

	if err := os.WriteFile(downFile, downJSON, 0644); err != nil {
		log.Fatalf("Failed to create down migration file: %v", err)
	}

	log.Printf("Created JSON migration files:\n  %s\n  %s", upFile, downFile)
}

func checkMigrationStatus(m *migrate.Migrate) {
	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			log.Println("No migrations have been applied yet")
			return
		}
		log.Fatalf("Failed to get migration status: %v", err)
	}

	log.Printf("Current migration version: %d", version)
	if dirty {
		log.Println("Database is in dirty state (migration failed)")
	} else {
		log.Println("Database is clean")
	}
}

func getMigrationVersion(m *migrate.Migrate) {
	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			fmt.Println("0")
			return
		}
		log.Fatalf("Failed to get migration version: %v", err)
	}

	fmt.Printf("%d", version)
	if dirty {
		fmt.Println(" (dirty)")
	}
}

func printUsage() {
	fmt.Println("Usage: go run ./cmd/migrate/migrate.go <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up                    Run all pending migrations")
	fmt.Println("  down                  Rollback the last migration")
	fmt.Println("  goto <version>        Migrate to specific version")
	fmt.Println("  force <version>       Force migration to specific version")
	fmt.Println("  create <name>         Create a new migration file")
	fmt.Println("  status                Show current migration status")
	fmt.Println("  version               Show current migration version")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -uri string           MongoDB connection URI (from .env)")
	fmt.Println("  --format string       Migration format: js or json (default: js)")
	fmt.Println()
	fmt.Println("Required Environment Variables (.env file):")
	fmt.Println("  MONGO_URI             MongoDB connection URI")
	fmt.Println("  MIGRATIONS_PATH       Path to migration files")
	fmt.Println()
	fmt.Println("Example .env file:")
	fmt.Println("  MONGO_URI=mongodb://localhost:27017/media_library")
	fmt.Println("  MIGRATIONS_PATH=deployments/mongo/migration")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run ./cmd/migrate/migrate.go up")
	fmt.Println("  go run ./cmd/migrate/migrate.go create add_users_collection")
	fmt.Println("  go run ./cmd/migrate/migrate.go create add_users_collection --format=json")
	fmt.Println("  go run ./cmd/migrate/migrate.go goto 2")
	fmt.Println("  go run ./cmd/migrate/migrate.go status")
	fmt.Println("  go run ./cmd/migrate/migrate.go down")
}
