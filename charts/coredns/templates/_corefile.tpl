{{- define "coredns.corefile" -}}
. {
  cache
  errors
  file /etc/coredns/Zonefile
  filterlist {
    blocklists https://adguardteam.github.io/AdGuardSDNSFilter/Filters/filter.txt
  }
  forward . tls://1.1.1.1 tls://1.0.0.1 {
    tls_servername cloudflare-dns.com
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
