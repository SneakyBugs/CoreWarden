{{- define "coredns.corefile" }}
. {
  slog
  errors
  filterlist {
    blocklists https://adguardteam.github.io/AdGuardSDNSFilter/Filters/filter.txt
  }
  forward . tls://1.1.1.1 tls://1.0.0.1 {
    tls_servername cloudflare-dns.com
    health_check 5s
  }
  health
  ready
}
{{- end }}
