{{/*
Expand the name of the chart.
*/}}
{{- define "supabase-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "supabase-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "supabase-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "supabase-operator.labels" -}}
helm.sh/chart: {{ include "supabase-operator.chart" . }}
{{ include "supabase-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "supabase-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "supabase-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "supabase-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "supabase-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Webhook service name
*/}}
{{- define "supabase-operator.webhookServiceName" -}}
{{- printf "%s-webhook" (include "supabase-operator.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Webhook certificate secret name
*/}}
{{- define "supabase-operator.webhookCertSecret" -}}
{{- printf "%s-webhook-cert" (include "supabase-operator.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}
