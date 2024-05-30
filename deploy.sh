#!/bin/bash

# Function to run the development environment
ENV_DEV="deploy/monolithic/.env.dev"
ENV_PROD="deploy/monolithic/.env.prod"

run_dev() {
ENV_FILE="deploy/monolithic/.env.dev"
docker compose -f docker_compose.yaml --env-file "$ENV_FILE" up
}

run_prod() {
// TODO ADD .env.prod
ENV_FILE="deploy/monolithic/.env.prod"
docker compose -f docker_compose.yaml --env-file "$ENV_FILE" up
}

remove_all() {
docker stop $(docker ps -aq)
docker rm $(docker ps -aq)
docker volume rm $(docker volume ls -qf dangling=true)
}



# Check if an argument is provided
if [ -z "$1" ]; then
    echo "Usage: $0 [run_dev|run_prod]"
    exit 1
fi

$@