{{- define "coredns.corefile" }}
. {
  debug
  errors
  filterlist
  forward . tls://1.1.1.1 tls://1.0.0.1 {
    tls_servername cloudflare-dns.com
    health_check 5s
  }
  health
  ready
}
{{- end }}
