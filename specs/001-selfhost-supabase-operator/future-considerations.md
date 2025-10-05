# Future Considerations - Supabase Operator

This document tracks important architectural decisions that are deliberately **deferred** for future iterations. The current design (v1alpha1) maintains flexibility to accommodate these features without breaking changes.

## Deferred to Future Iterations

### 1. Helm Chart Design 🎯
**Status**: Not in v1alpha1 scope
**Considerations**:
- Chart structure for operator deployment
- Configurable values (image tags, replicas, resources)
- Default values strategy
- Upgrade hooks and policies

**Why Deferred**:
- Simple manifest deployment sufficient for initial release
- Helm adds complexity to testing and distribution
- Can add later without CRD changes

**Compatibility Strategy**:
- Keep deployment simple (standard K8s manifests)
- Ensure all operator config via environment variables or flags
- Makes future Helm chart addition straightforward

---

### 2. RBAC Permissions 🔒
**Status**: Basic permissions only in v1alpha1
**Considerations**:
- Fine-grained RBAC scoping
- Namespace-scoped vs cluster-scoped operator deployment
- Secret management permissions (only manage operator-created secrets?)
- Multi-tenant RBAC isolation

**Why Deferred**:
- Start with broad permissions, tighten later
- Security hardening is iterative
- Difficult to predict all permission needs upfront

**Compatibility Strategy**:
- Document all RBAC requirements clearly
- Use ServiceAccount per operator instance
- Future versions can only reduce permissions (non-breaking)

---

### 3. Monitoring & Metrics 📊
**Status**: Basic metrics only in v1alpha1
**Considerations**:
- Complete Prometheus metrics catalog
- Grafana dashboard templates
- Alerting rules and SLOs
- Distributed tracing integration
- Custom metrics per component

**Why Deferred**:
- Metrics evolve based on operational experience
- Need production usage to identify important signals
- Over-engineering metrics creates maintenance burden

**Compatibility Strategy**:
- Use controller-runtime metrics infrastructure
- Follow Prometheus naming conventions
- Metrics are additive (no breaking changes)

---

### 4. Network Policies 🌐
**Status**: Not managed by operator in v1alpha1
**Considerations**:
- Automatic NetworkPolicy creation
- Default-deny approach
- Service mesh integration (Istio, Linkerd)
- Ingress controller integration
- mTLS between components

**Why Deferred**:
- Network requirements vary by cluster setup
- Users may have existing network policies
- Service mesh choice is deployment-specific

**Compatibility Strategy**:
- Use standard Kubernetes labels for all resources
- Label-based selectors enable external NetworkPolicy creation
- Component communication patterns well-documented
- Future operator can add NetworkPolicy CRs without spec changes

---

### 5. Upgrade Strategy 🔄
**Status**: Basic rolling updates only in v1alpha1
**Considerations**:
- Operator self-upgrade automation
- Blue-green deployment support
- Canary releases
- Rollback automation
- Version compatibility matrix
- CRD version migration tooling

**Why Deferred**:
- Complex upgrade scenarios need real-world testing
- Rollback conflicts with reconciliation pattern
- Need stability before automating upgrades

**Compatibility Strategy**:
- Follow semantic versioning strictly
- v1alpha1 → v1beta1 → v1 progression
- Implement conversion webhooks early
- Document breaking changes clearly

---

### 6. High Availability 🏗️
**Status**: Single operator replica in v1alpha1
**Considerations**:
- Operator leader election
- Multi-replica operator deployment
- Component anti-affinity rules
- PodDisruptionBudgets
- Cross-zone deployment
- Regional failover

**Why Deferred**:
- Adds complexity to initial deployment
- HA requirements vary by use case
- Needs production testing to validate

**Compatibility Strategy**:
- Use controller-runtime leader election (ready for HA)
- Component configs support replicas field
- Future HA features are configuration, not API changes

---

### 7. Missing Features ❓
**Status**: Explicitly out of scope for v1alpha1

#### 7.1 Backup & Restore
- Automated backup scheduling
- Point-in-time recovery
- Cross-region backup replication
- Backup validation and testing

**Defer Reason**: Complex feature needing separate design
**Future Path**: Separate CRD or integration with backup tools

#### 7.2 Migration Tooling
- Docker Compose → Kubernetes migration
- Import from existing Supabase installations
- Zero-downtime migration

**Defer Reason**: Needs production deployment experience
**Future Path**: Separate CLI tool or runbook

#### 7.3 Edge Functions Support
- Deno runtime deployment
- Function versioning
- Cold start optimization

**Defer Reason**: Marked optional in PRD
**Future Path**: Add when core components stable

#### 7.4 Multi-Cluster Support
- Cross-cluster replication
- Global routing
- Disaster recovery across regions

**Defer Reason**: Advanced use case
**Future Path**: Requires federation design

#### 7.5 Service Mesh Integration
- Istio/Linkerd integration
- mTLS automation
- Traffic splitting
- Circuit breaking

**Defer Reason**: Deployment environment-specific
**Future Path**: Documentation and examples, not operator code

---

## Architectural Flexibility Principles

The v1alpha1 design maintains flexibility through:

1. **Extensible CRD Structure**
   - Component configs use pointers (nil = defaults)
   - ExtraEnv allows arbitrary configuration
   - Status has room for additional fields

2. **Standard Kubernetes Patterns**
   - Uses corev1 types (compatible with ecosystem)
   - Follows Kubebuilder conventions
   - Label selectors enable external tooling

3. **Loose Coupling**
   - External dependencies (PostgreSQL, S3) not managed
   - No assumptions about ingress controller
   - No assumptions about storage class
   - No assumptions about network policies

4. **Additive Evolution**
   - New fields can be added without breaking changes
   - New condition types are additive
   - New metrics are additive
   - New RBAC permissions can be tightened

5. **Version Progression**
   - v1alpha1 signals experimental
   - Conversion webhooks planned for v1beta1
   - Breaking changes allowed in alpha

---

## Review Schedule

**Revisit after**:
- 3 months production usage
- 10+ deployments in different environments
- Community feedback collection
- Stability of v1alpha1 API

**Trigger Events**:
- Feature requests from multiple users
- Operational pain points identified
- Security vulnerabilities discovered
- Performance bottlenecks observed

---

## Implementation Guidance

For each deferred decision:
1. ✅ Document current limitations
2. ✅ Provide workarounds if possible
3. ✅ Reference this document in relevant code comments
4. ✅ Keep API flexible for future addition
5. ❌ Don't over-engineer for hypothetical use cases

---

*Last Updated*: 2025-10-03
*Review Frequency*: Quarterly after v1alpha1 release