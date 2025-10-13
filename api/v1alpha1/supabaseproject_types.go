package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SupabaseProjectSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	ProjectID string `json:"projectId"`

	// +kubebuilder:validation:Required
	Database DatabaseConfig `json:"database"`

	// +kubebuilder:validation:Required
	Storage StorageConfig `json:"storage"`

	// +optional
	Kong *KongConfig `json:"kong,omitempty"`

	// +optional
	Auth *AuthConfig `json:"auth,omitempty"`

	// +optional
	Realtime *RealtimeConfig `json:"realtime,omitempty"`

	// +optional
	PostgREST *PostgRESTConfig `json:"postgrest,omitempty"`

	// +optional
	StorageAPI *StorageAPIConfig `json:"storageApi,omitempty"`

	// +optional
	Meta *MetaConfig `json:"meta,omitempty"`

	// +optional
	Studio *StudioConfig `json:"studio,omitempty"`

	// +optional
	Ingress *IngressConfig `json:"ingress,omitempty"`
}

type DatabaseConfig struct {
	// SecretRef references a Secret containing PostgreSQL connection credentials.
	//
	// IMPORTANT: This operator currently only supports supabase/postgres image.
	// The Secret must contain the following keys:
	//   - host: PostgreSQL host (e.g., postgres.default.svc.cluster.local)
	//   - port: PostgreSQL port (e.g., 5432)
	//   - database: Database name (must be "postgres")
	//   - username: PostgreSQL user (must have SUPERUSER or be supabase_admin)
	//   - password: PostgreSQL password
	//
	// The user must have sufficient privileges to:
	//   - CREATE DATABASE (for _supabase database)
	//   - CREATE ROLE (for Supabase service roles)
	//   - CREATE EXTENSION (for pg_net and other extensions)
	//   - CREATE EVENT TRIGGER (requires superuser)
	//
	// Recommended: Use the "postgres" or "supabase_admin" user from supabase/postgres image.
	//
	// +kubebuilder:validation:Required
	SecretRef corev1.SecretReference `json:"secretRef"`

	// +kubebuilder:default="require"
	// +optional
	SSLMode string `json:"sslMode,omitempty"`

	// +kubebuilder:default=20
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	// +optional
	MaxConnections int `json:"maxConnections,omitempty"`
}

type StorageConfig struct {
	// +kubebuilder:validation:Required
	SecretRef corev1.SecretReference `json:"secretRef"`

	// +kubebuilder:default=true
	// +optional
	ForcePathStyle bool `json:"forcePathStyle,omitempty"`
}

type KongConfig struct {
	// +kubebuilder:default="kong:2.8.1"
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	ExtraEnv []corev1.EnvVar `json:"extraEnv,omitempty"`
}

type AuthConfig struct {
	// +kubebuilder:default="supabase/gotrue:v2.177.0"
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	SMTPSecretRef *corev1.SecretReference `json:"smtpSecretRef,omitempty"`

	// +optional
	OAuthSecretRef *corev1.SecretReference `json:"oauthSecretRef,omitempty"`

	// +optional
	ExtraEnv []corev1.EnvVar `json:"extraEnv,omitempty"`
}

type RealtimeConfig struct {
	// +kubebuilder:default="supabase/realtime:v2.34.47"
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	ExtraEnv []corev1.EnvVar `json:"extraEnv,omitempty"`
}

type PostgRESTConfig struct {
	// +kubebuilder:default="postgrest/postgrest:v12.2.12"
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	ExtraEnv []corev1.EnvVar `json:"extraEnv,omitempty"`
}

type StorageAPIConfig struct {
	// +kubebuilder:default="supabase/storage-api:v1.25.7"
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	ExtraEnv []corev1.EnvVar `json:"extraEnv,omitempty"`
}

type MetaConfig struct {
	// +kubebuilder:default="supabase/postgres-meta:v0.91.0"
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	ExtraEnv []corev1.EnvVar `json:"extraEnv,omitempty"`
}

type StudioConfig struct {
	// +kubebuilder:default="supabase/studio:2025.10.01-sha-8460121"
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	ExtraEnv []corev1.EnvVar `json:"extraEnv,omitempty"`

	// +optional
	PublicURL string `json:"publicUrl,omitempty"`

	// DashboardBasicAuthSecretRef references a Secret containing username/password
	// used to protect the Supabase Studio route behind Kong basic-auth. The Secret
	// must define the keys "username" and "password". When provided, Kong will
	// render the dashboard route with basic-auth enabled.
	// +optional
	DashboardBasicAuthSecretRef *corev1.SecretReference `json:"dashboardBasicAuthSecretRef,omitempty"`
}

type IngressConfig struct {
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// +optional
	ClassName *string `json:"className,omitempty"`

	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// +optional
	Host string `json:"host,omitempty"`

	// +optional
	TLSSecretName string `json:"tlsSecretName,omitempty"`
}

type SupabaseProjectStatus struct {
	// +optional
	Phase string `json:"phase,omitempty"`

	// +optional
	Message string `json:"message,omitempty"`

	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	Components ComponentsStatus `json:"components,omitempty"`

	// +optional
	Dependencies DependenciesStatus `json:"dependencies,omitempty"`

	// +optional
	Endpoints EndpointsStatus `json:"endpoints,omitempty"`

	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`
}

type ComponentsStatus struct {
	// +optional
	Kong ComponentStatus `json:"kong,omitempty"`

	// +optional
	Auth ComponentStatus `json:"auth,omitempty"`

	// +optional
	Realtime ComponentStatus `json:"realtime,omitempty"`

	// +optional
	PostgREST ComponentStatus `json:"postgrest,omitempty"`

	// +optional
	StorageAPI ComponentStatus `json:"storageApi,omitempty"`

	// +optional
	Meta ComponentStatus `json:"meta,omitempty"`

	// +optional
	Studio ComponentStatus `json:"studio,omitempty"`
}

type ComponentStatus struct {
	// +optional
	Phase string `json:"phase,omitempty"`

	// +optional
	Ready bool `json:"ready,omitempty"`

	// +optional
	Version string `json:"version,omitempty"`

	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`
}

type DependenciesStatus struct {
	// +optional
	PostgreSQL DependencyStatus `json:"postgresql,omitempty"`

	// +optional
	S3 DependencyStatus `json:"s3,omitempty"`
}

type DependencyStatus struct {
	Connected bool `json:"connected"`

	// +optional
	LastConnectedTime *metav1.Time `json:"lastConnectedTime,omitempty"`

	// +optional
	Error string `json:"error,omitempty"`

	// +optional
	LatencyMs int32 `json:"latencyMs,omitempty"`
}

type EndpointsStatus struct {
	// +optional
	API string `json:"api,omitempty"`

	// +optional
	Auth string `json:"auth,omitempty"`

	// +optional
	Realtime string `json:"realtime,omitempty"`

	// +optional
	Storage string `json:"storage,omitempty"`

	// +optional
	REST string `json:"rest,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SupabaseProject is the Schema for the supabaseprojects API
type SupabaseProject struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of SupabaseProject
	// +required
	Spec SupabaseProjectSpec `json:"spec"`

	// status defines the observed state of SupabaseProject
	// +optional
	Status SupabaseProjectStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// SupabaseProjectList contains a list of SupabaseProject
type SupabaseProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SupabaseProject `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SupabaseProject{}, &SupabaseProjectList{})
}
