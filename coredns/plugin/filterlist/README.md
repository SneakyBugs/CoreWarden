## Description

With _filterlist_ you get ad blocking in CoreDNS.

- Supports plain domain lists, hosts files, and adblock style blocklists.
- Robust blocklist fetching with retry, backoff, and stale blocklist refetching.

## Syntax

```
filterlist {
  blocklists URL...
}
```

- `blocklists` **URL...** links to filter lists of plain domains, hosts, and
  adblock style rules to be blocked.

## Metrics

- `coredns_filterlist_list_fetch_backoffs` - count of list fetch backoffs.
- `coredns_filterlist_list_fetch_failures` - count of list fetch failures.
- `coredns_filterlist_list_fetches_total` - count of total list fetches.
- `coredns_filterlist_requests_blocked` - count of blocked queries.
- `coredns_filterlist_requests_total` - count of handled queries, useful because this plugin runs behind `cache`.

## Examples

Use to block ads with Adguard filter list.

```
. {
  filterlist {
    blocklists https://adguardteam.github.io/AdGuardSDNSFilter/Filters/filter.txt
  }
  forward . tls://1.1.1.1 tls://1.0.0.1 {
    tls_servername cloudflare-dns.com
    health_check 5s
  }
}
```
