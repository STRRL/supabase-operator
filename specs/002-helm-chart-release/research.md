# Research: Automated Helm Chart Release

**Feature**: 002-helm-chart-release | **Date**: 2025-11-23

## Key Decisions

### 1. GitHub Action Selection
**Decision**: Use `stefanprodan/helm-gh-pages` action
**Rationale**: Mature, well-maintained action specifically designed for Helm charts on GitHub Pages. Handles all packaging, versioning, and index updates automatically.
**Alternatives considered**:
- `helm/chart-releaser-action`: More complex, designed for GitHub Releases rather than Pages
- Custom script: Unnecessary complexity when existing action meets all requirements

### 2. Workflow Trigger Strategy
**Decision**: Trigger on semantic version tags (v*.*.*)
**Rationale**: Aligns with spec requirement FR-001, ensures deliberate releases, prevents accidental publishing
**Alternatives considered**:
- Push to main branch: Too frequent, would publish incomplete work
- Manual dispatch: Requires human intervention, defeats automation purpose

### 3. Authentication Method
**Decision**: GitHub Personal Access Token (PAT) stored as secret
**Rationale**: Specified in clarifications, provides necessary permissions for cross-repository push
**Alternatives considered**:
- GITHUB_TOKEN: Insufficient permissions for pushing to external repository
- Deploy keys: More complex setup, PAT is simpler and sufficient

### 4. Chart Repository Structure
**Decision**: Host on GitHub Pages at https://github.com/STRRL/helm.strrl.dev
**Rationale**: Free hosting, integrated with GitHub, standard approach for Helm charts
**Alternatives considered**:
- ChartMuseum: Requires hosting infrastructure
- ArtifactHub only: Doesn't host charts, only indexes them

### 5. Version Extraction
**Decision**: Extract version from git tag and apply to Chart.yaml
**Rationale**: Single source of truth for versioning, tag-driven releases
**Alternatives considered**:
- Manual Chart.yaml updates: Prone to inconsistency
- Automated commits: Creates noise in git history

### 6. Concurrency Control
**Decision**: Implement workflow concurrency group
**Rationale**: Prevents race conditions during simultaneous releases, ensures index.yaml integrity
**Alternatives considered**:
- No control: Risk of corrupted repository index
- Queue all runs: Implemented via cancel-in-progress: false

### 7. Error Handling
**Decision**: Fail fast with no notifications (per spec clarification)
**Rationale**: Aligns with user preference, GitHub Actions UI shows failures
**Alternatives considered**:
- Email notifications: User explicitly chose "No Notification"
- Slack webhooks: Unnecessary complexity

## Best Practices Applied

### GitHub Actions Configuration
- Use `actions/checkout@v4` with `fetch-depth: 0` for full history
- Pin action versions for reproducibility
- Set explicit permissions (`contents: write`)
- Configure git user for commits

### Helm Chart Publishing
- Validate charts with `helm lint` before publishing
- Maintain all chart versions in repository
- Generate proper index.yaml with chart URLs
- Preserve existing versions (no overwrites)

### Security Measures
- Store PAT as encrypted secret
- Use minimum required permissions
- Never expose tokens in logs
- Consider chart signing for production (future enhancement)

### Repository Structure
```
.github/workflows/release-helm.yaml  # Workflow definition
charts/supabase-operator/            # Existing chart location
  ├── Chart.yaml                     # Version updated by workflow
  ├── values.yaml
  └── templates/
```

### Target Repository Structure (gh-pages)
```
index.yaml                           # Repository index
supabase-operator-0.1.0.tgz         # Packaged charts
supabase-operator-0.2.0.tgz
...
```

## Required GitHub Configuration

### 1. Personal Access Token Setup
- Create PAT with `repo` scope (for public repository)
- Add as secret `HELM_REPO_TOKEN` in source repository
- Token needs write access to STRRL/helm.strrl.dev repository

### 2. Target Repository Setup
- Repository `STRRL/helm.strrl.dev` must exist
- Enable GitHub Pages from `gh-pages` branch
- Initialize gh-pages branch if not present

### 3. Workflow Permissions
- Ensure Actions have write permissions in source repository
- Settings → Actions → General → Workflow permissions → Read and write

## Implementation Approach

### Phase 1: Core Workflow
1. Create `.github/workflows/release-helm.yaml`
2. Configure tag pattern matching (v*.*.*)
3. Extract version from tag
4. Update Chart.yaml versions
5. Package and publish using helm-gh-pages action

### Phase 2: Validation & Testing
1. Add helm lint step
2. Validate chart installation in test namespace
3. Verify repository index generation
4. Test concurrent release handling

### Phase 3: Documentation
1. Update README with installation instructions
2. Document release process
3. Add troubleshooting guide

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| PAT expiration | Release failures | Set expiration reminder, document rotation |
| Concurrent releases | Corrupted index | Concurrency control in workflow |
| Invalid chart syntax | Failed installation | Helm lint validation |
| Target repo unavailable | No release | Fail fast, rely on GitHub status |

## Future Enhancements

1. **Chart Signing**: Implement GPG signing for production readiness
2. **Automated Testing**: Add chart installation tests in Kind cluster
3. **Release Notes**: Generate changelog from commit messages
4. **ArtifactHub Integration**: Register chart for better discoverability
5. **Multi-Chart Support**: Extend to support multiple charts if needed

## References

- [stefanprodan/helm-gh-pages documentation](https://github.com/stefanprodan/helm-gh-pages)
- [Helm Chart Repository Guide](https://helm.sh/docs/topics/chart_repository/)
- [GitHub Actions workflow syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
- [Reference workflow](https://github.com/STRRL/cloudflare-tunnel-ingress-controller/blob/master/.github/workflows/release-helm.yaml)