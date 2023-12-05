{{- define "coredns.corefile" -}}
. {
  cache
  errors
  {{- with .Values.config.zones }}
  file /etc/coredns/Zonefile{{ range $zone, $_ := . }} {{ $zone }}{{ end }}
  {{- end }}
  filterlist {
    blocklists{{ range .Values.config.filter.blocklists }} {{ . }}{{ end }}
  }
  forward .{{ range .Values.config.upstream.servers }} {{ . }}{{ end }} {
    tls_servername {{ .Values.config.upstream.name }}
    health_check 5s
  }
  health
  prometheus 0.0.0.0:9153
  ready
  slog
}
{{- end }}
{{- define "coredns.zonefile" }}
{{- with .Values.config.zones }}
{{- range $zone, $records := . -}}
{{- if not (hasSuffix "." $zone) }}
{{- fail (printf "\nZone name must end with a dot, found '%s' inside .Values.config.zones\nreplace it with '%s.'" $zone $zone) }}
{{- end }}
$ORIGIN {{ $zone }}
{{- /* Dummy SOA record so users wont have to manually configure it. */}}
@ 0 IN SOA ns.icann.org. noc.dns.icann.org. 2020091001 7200 3600 1209600 3600
{{- range $records }}
{{ . }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
