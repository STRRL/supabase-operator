# Product Requirements Document: Supabase Operator for Kubernetes

## Problem

Currently, there is no official Supabase operator for Kubernetes. Organizations wanting to deploy Supabase on Kubernetes must manually manage the deployment and configuration of all components, which is complex and error-prone.

Additionally, self-hosted Supabase does not support multiple projects very well. Each Supabase instance is designed to handle a single project, making it difficult for organizations to manage multiple isolated projects efficiently on Kubernetes without deploying completely separate infrastructure for each one.

### Component Analysis

**Supabase Native Services** that need to be managed:

1. **Kong (API Gateway)** (`supabase/kong`)
   - API gateway and router
   - Handles authentication and routing

2. **GoTrue (Auth)** (`supabase/gotrue`)
   - Authentication service
   - User management and JWT tokens

3. **PostgREST** (`postgrest/postgrest`)
   - Auto-generated REST API from PostgreSQL schema
   - Core component for database access

4. **Realtime** (`supabase/realtime`)
   - WebSocket server for real-time subscriptions
   - Broadcasts database changes

5. **Storage API** (`supabase/storage-api`)
   - Handles file storage operations
   - Requires PostgreSQL database
   - Requires object storage backend (S3-compatible)
   - Supports image transformation via imgproxy

6. **Meta (pg-meta)** (`supabase/postgres-meta`)
   - PostgreSQL metadata API
   - Database schema management

7. **Functions (Edge Functions)** (optional initially)
   - Serverless function runtime
   - Deno-based execution environment

**Third-Party Dependencies** that Supabase depends on:

1. **MinIO** (`minio/minio`)
   - S3-compatible object storage
   - Used as storage backend for files
   - Requires bucket initialization

2. **imgproxy** (`darthsim/imgproxy`)
   - Image transformation service
   - Optional but recommended for storage features

3. **PostgreSQL** (referenced as `db` service)
   - Primary database
   - Core dependency for all Supabase services

## Proposal

Build a minimal viable Supabase operator that leverages existing Kubernetes operators for third-party dependencies.

### Use Existing Operators

1. **PostgreSQL**: Use existing PostgreSQL operators
   - CloudNativePG
   - Zalando Postgres Operator
   - Crunchy Data PostgreSQL Operator

2. **Object Storage**: Use existing MinIO operator or S3-compatible services
   - MinIO Operator
   - Or integrate with cloud providers (AWS S3, GCS, Azure Blob)

3. **Image Processing**: Deploy imgproxy as standard Deployment/StatefulSet
   - No specialized operator needed
   - Simple stateless service

### Operator Responsibilities

The Supabase operator should focus on:

1. **Supabase Native Services Management**
   - Deploy and configure Kong API Gateway
   - Deploy and configure GoTrue (Auth service)
   - Deploy and configure PostgREST
   - Deploy and configure Realtime
   - Deploy and configure Storage API
   - Deploy and configure Meta (pg-meta)
   - Optionally deploy Functions runtime

2. **Integration & Orchestration**
   - Coordinate between Supabase services and external dependencies
   - Manage service discovery and configuration
   - Handle secrets and credentials management

3. **Configuration Management**
   - JWT secrets
   - API keys (ANON_KEY, SERVICE_ROLE_KEY)
   - Database connection strings
   - Storage backend configuration

### Dependencies Management Modes

1. **Managed Mode**: Operator creates and manages all dependencies
   - Deploy PostgreSQL using external operator
   - Deploy MinIO using external operator
   - Deploy imgproxy as Deployment

2. **External Mode**: User provides existing services
   - Accept connection strings for external PostgreSQL
   - Accept S3 configuration for external object storage
   - Optional external imgproxy

### Success Criteria

1. Deploy a complete working Supabase instance with ALL core services in one command
2. Support all major Supabase features (Auth, Database REST API, Realtime, Storage, Metadata)
3. Integrate seamlessly with existing PostgreSQL and object storage solutions
4. Handle configuration updates and rolling upgrades
5. Provide clear status reporting on resource health
6. Support both managed and external dependency modes
