{{/*
Expand the name of the chart.
*/}}
{{- define "crm-platform.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "crm-platform.fullname" -}}
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
{{- define "crm-platform.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "crm-platform.labels" -}}
helm.sh/chart: {{ include "crm-platform.chart" . }}
{{ include "crm-platform.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: crm-platform
{{- end }}

{{/*
Selector labels
*/}}
{{- define "crm-platform.selectorLabels" -}}
app.kubernetes.io/name: {{ include "crm-platform.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Service account name
*/}}
{{- define "crm-platform.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "crm-platform.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Image name helper
*/}}
{{- define "crm-platform.image" -}}
{{- $registry := .global.imageRegistry | default "ghcr.io" }}
{{- printf "%s/%s:%s" $registry .image.repository (.image.tag | default "latest") }}
{{- end }}

{{/*
Generate service labels
*/}}
{{- define "crm-platform.serviceLabels" -}}
app.kubernetes.io/name: {{ .name }}
app.kubernetes.io/component: backend
app.kubernetes.io/part-of: crm-platform
{{- end }}
