{{- define "coredns.corefile" -}}
. {
  cache
  errors
  file /etc/coredns/Zonefile
  filterlist {
    blocklists{{ range .Values.config.filter.blocklists }} {{ . }}{{ end }}
  }
  forward .{{ range .Values.config.upstream.servers }} {{ . }}{{ end }} {
    tls_servername {{ .Values.config.upstream.name }}
    health_check 5s
  }
  health
  prometheus
  ready
  slog
}
{{- end }}
{{- define "coredns.zonefile" }}
{{- with .Values.config.zones }}
{{- range $zone, $records := . -}}
$ORIGIN {{ $zone }}
{{- /* Dummy SOA record so users wont have to manually configure it. */}}
@ 3600 IN SOA ns.icann.org. noc.dns.icann.org. 2020091001 7200 3600 1209600 3600
{{- range $records }}
{{ . }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
