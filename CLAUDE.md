# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a Grafana backend data source plugin for Google Cloud Logging. It allows users to query and visualize Google Cloud logs within Grafana dashboards.

## Key Commands

### Backend Development (Go)
```bash
# Build the backend plugin
mage -v

# Run backend tests
go test ./pkg/...

# Run backend tests with coverage
go test -coverprofile=coverage.out ./pkg/...
go tool cover -html=coverage.out
```

### Frontend Development (TypeScript/React)
```bash
# Install dependencies
yarn install

# Development mode with hot reload
yarn dev

# Production build
yarn build

# Run frontend tests
yarn test

# Run tests in watch mode
yarn test --watch
```

### Full Plugin Build
```bash
# Build both frontend and backend
mage -v && yarn build
```

## Architecture Overview

### Backend Structure (Go)
- **Entry Point**: `pkg/main.go` - Initializes the Grafana plugin
- **Core Plugin Logic**: `pkg/plugin/plugin.go` - Implements datasource interface for query handling and health checks
- **Cloud Logging Client**: `pkg/plugin/cloudlogging/` - Handles Google Cloud Logging API interactions
  - `client.go`: API client implementation with authentication
  - `cloudlogging.go`: Core logging functionality and query processing
- **Authentication**: Supports JWT (service account key) and GCE default service account

### Frontend Structure (TypeScript/React)
- **Datasource Class**: `src/datasource.ts` - Frontend query handling and variable support
- **Configuration UI**: `src/ConfigEditor.tsx` - Authentication and project setup
- **Query Editor**: `src/QueryEditor.tsx` - UI for building Cloud Logging queries
- **Variable Support**: `src/VariableQueryEditor.tsx` - Template variable queries for projects/buckets/views

### Key Integration Points
1. **Authentication Flow**: Frontend ConfigEditor → Backend plugin → Google Cloud APIs
2. **Query Execution**: QueryEditor → datasource.ts → Backend plugin.go → cloudlogging.go → Google Cloud Logging API
3. **Variable Resolution**: VariableQueryEditor → Backend metadata endpoints

## Testing Approach

### Backend Testing
- Unit tests alongside implementation files (`*_test.go`)
- Mock interfaces in `pkg/plugin/mocks/` for testing Cloud Logging client
- Use `testify` for assertions and mocking

### Frontend Testing
- Jest tests for TypeScript components
- Test files named `*.test.ts`
- Grafana Toolkit provides Jest configuration

## Required GCP Permissions

The service account needs:
- **Logs Viewer** role
- **Logs View Accessor** role (for log scopes)
- **Cloud Resource Manager API** must be enabled (for project listing)
- **Service Account Token Creator** role (only if using service account impersonation)

## Development Tips

1. **Query Language**: Uses [Google Cloud Logging Query Language](https://cloud.google.com/logging/docs/view/logging-query-language)
2. **Log Scopes**: Support for projects, buckets, and views as template variables
3. **Time Range**: Grafana's time range is automatically applied to queries
4. **Annotations**: Supported through the query editor
5. **Alerting**: Not directly supported - use Log-based metrics with Cloud Monitoring instead

## Common Development Tasks

### Adding a New Query Feature
1. Update types in `src/types.ts`
2. Add UI controls in `src/QueryEditor.tsx`
3. Update query building logic in `src/datasource.ts`
4. Implement backend support in `pkg/plugin/cloudlogging/cloudlogging.go`

### Debugging Authentication Issues
1. Check service account permissions in GCP Console
2. Verify Cloud Resource Manager API is enabled
3. Test authentication in `pkg/plugin/plugin.go` CheckHealth method
4. Review logs for specific permission errors

### Running Locally with Grafana
1. Build the plugin: `mage -v && yarn build`
2. Copy/symlink dist folder to Grafana plugins directory
3. Configure datasource with service account JSON or use GCE authentication