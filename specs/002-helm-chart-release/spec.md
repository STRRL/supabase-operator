# Feature Specification: Automated Helm Chart Release

**Feature Branch**: `002-helm-chart-release`
**Created**: 2025-11-23
**Status**: Draft
**Input**: User description: "release helm charts to https://github.com/STRRL/helm.strrl.dev/tree/gh-pages, 参考https://github.com/STRRL/cloudflare-tunnel-ingress-controller/blob/master/.github/workflows/release-helm.yaml"

## Clarifications

### Session 2025-11-23

- Q: Which version tag patterns should trigger the release workflow? → A: Only v*.*.* semantic version tags (e.g., v1.2.3)
- Q: When a chart release fails during the publishing process, how should the system respond? → A: Alert maintainers and stop release
- Q: How should the system authenticate to push charts to the GitHub Pages repository? → A: GitHub personal access token (PAT)
- Q: What should be the name of the Helm chart that users will install? → A: supabase-operator
- Q: How should maintainers be alerted when a chart release fails? → A: No Notification

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Automated Chart Publishing on Version Tag (Priority: P1)

As a maintainer, when I create a new version tag for the Supabase Operator, the Helm charts should be automatically packaged and published to the central Helm repository, making them immediately available for users to install.

**Why this priority**: This is the core functionality that enables users to consume the operator through Helm. Without automated publishing, users cannot easily install or upgrade the operator, severely limiting adoption and usability.

**Independent Test**: Can be fully tested by pushing a version tag and verifying the chart appears in the Helm repository with correct version and is installable via standard Helm commands.

**Acceptance Scenarios**:

1. **Given** a new semantic version tag (e.g., v1.2.3) is pushed to the repository, **When** the release process triggers, **Then** the Helm chart is packaged with version 1.2.3 and published to the Helm repository
2. **Given** the Helm chart is published, **When** a user adds the repository and searches for charts, **Then** the new version appears in the available charts list
3. **Given** multiple version tags exist, **When** users list available versions, **Then** all published versions are displayed in chronological order

---

### User Story 2 - Chart Installation and Upgrade (Priority: P2)

As an end user, I want to easily install and upgrade the Supabase Operator using standard Helm commands, with all necessary resources properly configured and deployed.

**Why this priority**: While publishing is critical (P1), the user experience of actually installing and upgrading is what delivers the value. This ensures the published charts work correctly in real environments.

**Independent Test**: Can be tested by installing the chart in a test cluster and verifying all operator components deploy successfully and function as expected.

**Acceptance Scenarios**:

1. **Given** the Helm repository is added, **When** a user runs the install command, **Then** the operator and all required resources are deployed successfully
2. **Given** an existing installation, **When** a user upgrades to a new version, **Then** the upgrade completes without data loss or service disruption
3. **Given** installation parameters need customization, **When** users provide custom values, **Then** the deployment respects all provided configurations

---

### User Story 3 - Chart Discovery and Documentation (Priority: P3)

As a potential user, I want to discover available Helm charts, understand their capabilities, and access installation documentation through standard Helm repository interfaces.

**Why this priority**: Discovery and documentation enhance adoption but are not critical for basic functionality. The operator can function without perfect discoverability, though it impacts user experience.

**Independent Test**: Can be tested by accessing the Helm repository index and verifying metadata, descriptions, and links to documentation are present and accurate.

**Acceptance Scenarios**:

1. **Given** the Helm repository URL, **When** users access the index, **Then** they see chart metadata including description, version, and maintainer information
2. **Given** a published chart, **When** users examine the chart, **Then** they find comprehensive README with installation instructions and configuration options

---

### Edge Cases

- What happens when a tag is pushed that doesn't follow semantic versioning format v*.*.* (ignored, no release triggered)?
- How does the system handle concurrent releases from multiple tags pushed simultaneously?
- What occurs if the target repository is temporarily unavailable during publishing (stop release, no notification)?
- How are pre-release versions (e.g., v1.0.0-rc1) handled in the repository (not triggered per semantic version pattern)?
- What happens if chart validation fails due to malformed templates (stop release, no notification)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST automatically trigger chart packaging only when semantic version tags matching pattern v*.*.* are created
- **FR-002**: System MUST extract version numbers from git tags and apply them to chart versions
- **FR-003**: System MUST publish packaged charts to the designated repository location
- **FR-004**: Charts MUST be accessible via standard Helm repository protocols
- **FR-005**: System MUST maintain an updated index of all available chart versions
- **FR-006**: Each chart version MUST include all necessary templates and dependencies for operator deployment
- **FR-007**: System MUST preserve existing chart versions when publishing new ones
- **FR-008**: Charts MUST include comprehensive metadata (name, version, description, maintainers)
- **FR-009**: System MUST validate chart integrity before publishing
- **FR-010**: Published charts MUST be installable using standard Helm commands without additional configuration
- **FR-011**: System MUST stop the release process when any failure occurs during publishing (no notification required)
- **FR-012**: System MUST authenticate to the GitHub Pages repository using a GitHub personal access token (PAT)
- **FR-013**: The published Helm chart MUST be named "supabase-operator" in the repository index

### Key Entities

- **Helm Chart**: Package named "supabase-operator" containing all templates, values, and metadata needed to deploy the Supabase Operator
- **Chart Version**: Specific release of the chart, corresponding to a git tag, with semantic versioning
- **Chart Repository**: Centralized location where all chart versions are stored and indexed for distribution
- **Chart Index**: Metadata file listing all available charts and versions with their locations

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: New chart versions are available for installation within 5 minutes of tag creation
- **SC-002**: 100% of semantic version tags (v*.*.*) result in successfully published and installable charts
- **SC-003**: Users can install the operator with a single Helm command in under 30 seconds
- **SC-004**: Chart repository maintains 99.9% availability for chart downloads
- **SC-005**: All published charts pass validation checks without errors
- **SC-006**: Chart upgrades complete successfully 95% of the time without manual intervention
- **SC-007**: Repository index updates reflect new versions within 1 minute of publishing
