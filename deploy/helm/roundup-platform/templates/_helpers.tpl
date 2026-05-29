{{- define "roundup-platform.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "roundup-platform.fullname" -}}
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

{{- define "roundup-platform.labels" -}}
app.kubernetes.io/name: {{ include "roundup-platform.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "roundup-platform.image" -}}
{{- $registry := .Values.global.imageRegistry }}
{{- $repository := .image }}
{{- $tag := default .Values.global.imageTag .tag }}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- else }}
{{- printf "%s:%s" $repository $tag }}
{{- end }}
{{- end }}

{{- define "roundup-platform.env" -}}
- name: DATABASE_URL
  valueFrom:
    configMapKeyRef:
      name: roundup-config
      key: DATABASE_URL
- name: KAFKA_BROKERS
  valueFrom:
    configMapKeyRef:
      name: roundup-config
      key: KAFKA_BROKERS
{{- end }}

{{- define "roundup-platform.service-template" -}}
{{- $svc := .service -}}
{{- $name := .name -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ $name }}
  namespace: {{ $.Values.namespace }}
  labels:
    app: {{ $name }}
    {{- include "roundup-platform.labels" $ | nindent 4 }}
spec:
  replicas: {{ $svc.replicaCount }}
  selector:
    matchLabels:
      app: {{ $name }}
  template:
    metadata:
      labels:
        app: {{ $name }}
    spec:
      containers:
      - name: {{ $name }}
        image: {{ include "roundup-platform.image" (dict "image" $svc.image "tag" $.Values.global.imageTag) }}
        imagePullPolicy: {{ $.Values.global.imagePullPolicy }}
        ports:
        - containerPort: {{ $svc.port }}
        envFrom:
        - configMapRef:
            name: roundup-config
        {{- if $svc.env }}
        env:
        {{- range $key, $val := $svc.env }}
        {{- if and (eq $key "JWT_SECRET") (empty $val) }}
        - name: JWT_SECRET
          valueFrom:
            configMapKeyRef:
              name: roundup-config
              key: JWT_SECRET
        {{- else if $val }}
        - name: {{ $key }}
          value: {{ $val | quote }}
        {{- end }}
        {{- end }}
        {{- end }}
        {{- if $svc.livenessProbe }}
        livenessProbe:
          httpGet:
            path: {{ $svc.livenessProbe.path }}
            port: {{ $svc.livenessProbe.port }}
          initialDelaySeconds: 5
          periodSeconds: 10
        {{- end }}
        {{- if $svc.readinessProbe }}
        readinessProbe:
          httpGet:
            path: {{ $svc.readinessProbe.path }}
            port: {{ $svc.readinessProbe.port }}
          initialDelaySeconds: 3
          periodSeconds: 5
        {{- end }}
        resources:
          {{- toYaml $svc.resources | nindent 10 }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ $name }}
  namespace: {{ $.Values.namespace }}
  labels:
    app: {{ $name }}
spec:
  ports:
  - port: {{ $svc.port }}
    targetPort: {{ $svc.port }}
  selector:
    app: {{ $name }}
{{- end }}
