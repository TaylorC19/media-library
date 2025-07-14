# Database Migrations

This project uses a custom migration system built with Go and MongoDB. The migration tool is located in `cmd/migrate/migrate.go`.

## Prerequisites

1. MongoDB container running (use `docker-compose up` in the `deployments/docker/` directory)
2. Initialize the MongoDB container: `sh ./deployments/bin/dev_mongo_init.sh`

## Migration Commands

### Run Migrations
```bash
# Run all pending migrations
go run ./cmd/migrate/migrate.go up

# Run with custom MongoDB URI
go run ./cmd/migrate/migrate.go up -uri mongodb://localhost:27017 -db media_library
```

### Rollback Migrations
```bash
# Rollback the last migration
go run ./cmd/migrate/migrate.go down

# Rollback with custom settings
go run ./cmd/migrate/migrate.go down -uri mongodb://localhost:27017 -db media_library
```

### Create New Migration
```bash
# Create a new migration file
go run ./cmd/migrate/migrate.go create add_new_feature

# This creates:
# - deployments/mongo/migration/YYYYMMDDHHMMSS_add_new_feature.up.js
# - deployments/mongo/migration/YYYYMMDDHHMMSS_add_new_feature.down.js
```

### Check Migration Status
```bash
# Show current migration status
go run ./cmd/migrate/migrate.go status

# Show current version number
go run ./cmd/migrate/migrate.go version
```

### Force Migration
```bash
# Force migration to specific version (useful for fixing dirty state)
go run ./cmd/migrate/migrate.go force 1
```

## Migration File Format

Migrations are JavaScript files that run directly in MongoDB:

### Up Migration (.up.js)
```javascript
// Create collections, indexes, or insert data
db.createCollection("users");
db.users.createIndex({ "email": 1 }, { unique: true });
```

### Down Migration (.down.js)
```javascript
// Rollback operations
db.users.drop();
```

## File Structure

```
deployments/mongo/migration/
├── 001_initial_setup.up.js
├── 001_initial_setup.down.js
├── 002_add_feature.up.js
├── 002_add_feature.down.js
└── ...
```

## Configuration

Database settings can be configured via command line flags:
- `-uri`: MongoDB connection URI (default: mongodb://localhost:27017)
- `-db`: Database name (default: media_library)

## Integration with Main Application

The main application (`cmd/media-library/main.go`) automatically connects to MongoDB on startup using the same connection settings.

## Troubleshooting

1. **"No migrations to apply"**: This is normal when all migrations are already applied
2. **"Database is in dirty state"**: Use `force` command to fix
3. **Connection errors**: Ensure MongoDB container is running and accessible 