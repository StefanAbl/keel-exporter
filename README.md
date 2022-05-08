# keel-exporter
[![Create and publish a Docker image](https://github.com/StefanAbl/keel-exporter/actions/workflows/docker-image.yml/badge.svg)](https://github.com/StefanAbl/keel-exporter/actions/workflows/docker-image.yml)

Prometheus Exporter for Keel

## Metrics
The following metrics are exported

- keel_namespaces_total
- keel_pending_approvals_total
- keel_registries_total
- keel_tracked_images_total

## Configuration

The application is configured using environment variables: 

- KEEL_URL: The URL to the instance of Keel to monitor
- KEEL_USER: User name of the user configured for the Keel web interface
- KEEL_PASS: Password of the user configured for the Keel web interface
