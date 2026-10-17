---
title: API authorization
---

Authorization is controlled using `csv` policy files.
Each line in the policy file starts with `p` for a policy line, or `g` for a group line.

## Policies

Policy lines control what actions a subject (user or group) can perform on an object in a zone.
The policy line is structured in the following way:

```
p, subject, object, zone, action
```

For example a policy that allows `bob` to read and edit `records` looks like this:

```csv
p, bob, records, example.com., read
p, bob, records, example.com., edit
```

The `zones` object controls zone discovery.
`GET /v1/zones` lists only the configured zones the subject has `read` on:

```csv
p, bob, zones, example.com., read
```

The ExternalDNS webhook service account needs `zones` read on every zone it manages, in addition to `records` read and edit.

## Roles

Group lines define user presence in groups.
The group line is structured in the following way:

```
g, subject, group
```

For example the following policy file defines `alice` as a member of the `admins` group:

```csv
p, admins, records, example.com., read
g, alice, admins
```

## Policy file example

The following is a full policy file example with both policy and group definitions.

```csv
p, admins, records, example.com., read
p, admins, records, example.com., edit
g, alice, admins
g, bob, admins
p, carol, records, example.net., read
```
