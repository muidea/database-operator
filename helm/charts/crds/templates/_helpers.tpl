{{/* vim: set filetype=mustache: */}}

{{- define "database.crds.labels" -}}
{{- template "database.labels.merge" (list
  (include "database.labels.common" .)
  (include "database.crds.matchLabels" .)
  (toYaml .Values.customLabels)
) -}}
{{- end -}}

{{- define "database.crds.matchLabels" -}}
{{- template "database.labels.merge" (list
  (include "database.matchLabels.common" .)
  (include "database.labels.component" "crds")
) -}}
{{- end -}}
