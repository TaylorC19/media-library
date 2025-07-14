# media-library
Track your physical media collection for easy reference


## Getting Started
### development db
copy the `.env.sample` file to `.env`

Run the docker container
1. `docker compose -f deployments/docker/docker-compose.yaml up -d`
2. `sh deployments/bin/dev_mongo_init.sh`
3. `go run cmd/migrate/migrate.go up`