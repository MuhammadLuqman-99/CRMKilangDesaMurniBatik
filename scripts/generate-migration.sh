#!/bin/bash
# Generate Database Migration Script
# ===================================

set -e

if [ -z "$1" ]; then
    echo "Usage: ./generate-migration.sh <service> <migration_name>"
    echo "Services: iam, sales, notification"
    echo "Example: ./generate-migration.sh iam create_users_table"
    exit 1
fi

if [ -z "$2" ]; then
    echo "Please provide a migration name"
    exit 1
fi

SERVICE=$1
NAME=$2

case $SERVICE in
    iam|sales|notification)
        migrate create -ext sql -dir migrations/$SERVICE -seq $NAME
        echo "Migration created in migrations/$SERVICE/"
        ;;
    *)
        echo "Unknown service: $SERVICE"
        echo "Available services: iam, sales, notification"
        exit 1
        ;;
esac
