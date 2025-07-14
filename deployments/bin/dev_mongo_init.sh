#!/bin/bash

MONGO_CONTAINER_NAME="media_lib_mongo"

echo "# [RUN] dev_mongo_init.sh"

# Copy the init script to the container
docker cp deployments/mongo/init.sh ${MONGO_CONTAINER_NAME}:/tmp/init.sh

# Make the script executable and run it using bash
docker exec ${MONGO_CONTAINER_NAME} bash -c "chmod +x /tmp/init.sh && /tmp/init.sh"

# Clean up
docker exec ${MONGO_CONTAINER_NAME} rm -f /tmp/init.sh

echo "# [END] dev_mongo_init.sh"