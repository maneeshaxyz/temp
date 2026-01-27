#!/bin/bash
set -e

echo "Running initial cvd update..."
cvd update

echo "Initial update complete. Starting 2-hour update loop..."
while true; do
    sleep 7200  # 2 hours
    echo "Running cvd update at $(date)"
    cvd update
    echo "Update complete at $(date)"
done