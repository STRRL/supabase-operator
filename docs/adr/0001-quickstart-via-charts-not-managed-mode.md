# ADR 0001: Quickstart uses well known Helm charts, not an operator managed mode

Date: 2026-07-25

Status: accepted

## Context

Issue #18 planned a "managed mode": `spec.database.managed: true` would make the operator provision an in-cluster PostgreSQL StatefulSet, and a similar flag would provision MinIO. The goal was a first working project in under 5 minutes on a bare cluster.

Building that inside the operator means the operator takes on database lifecycle work: storage, upgrades, credentials, failure recovery. That is a large and permanent maintenance load, and mature tools already do it better.

A second constraint: the operator only supports the `supabase/postgres` image, because database initialization needs Supabase specific extensions (pgjwt, pg_net and others) and superuser access. Any install path must produce a PostgreSQL running that image.

## Decision

Do not build managed mode. Keep PostgreSQL and S3 strictly user provided.

Instead, ship a documented quickstart that installs the dependencies with well known Helm charts:

- PostgreSQL via the CloudNativePG operator, with the cluster image set to `supabase/postgres`. Known required settings: `postgresUID: 101`, `postgresGID: 102` (the image does not use the CNPG default uid 26), and `enableSuperuserAccess: true` (database initialization needs superuser).
- Object storage via the official MinIO chart (`charts.min.io`), standalone mode, bucket declared through chart values.

The quickstart manifests live in the repo so users can apply them directly and docs cannot drift from them.

## Consequences

- The operator stays thin: it never owns database or storage lifecycle.
- The 5 minute goal is met by documentation quality, not new operator code paths.
- The CNPG plus `supabase/postgres` combination is community proven but not officially supported by either project; the quickstart must be verified end to end whenever images are bumped.
- Issue #18 Phase A tasks about managed mode are superseded by this decision.
