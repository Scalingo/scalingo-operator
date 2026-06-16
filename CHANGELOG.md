# Changelog

## To Be Released

* feat(db/deletion) Support database resource deletion when the resource was deleted first through Scalingo API
* feat(iprange) Forward networking.ip_range attribute to the database creation params to choose the ip range

## v1.3.0

* feat(oks-net-peering) Add outscale:oks:net_peering parameter to automatically create oks net peering request acceptation and configure the database to use it
* build(deps): update to latest major version `github.com/Scalingo/go-scalingo/v11`

## v1.2.0

* feat: set Scalingo Operator User-Agent on go-scalingo API calls
* refactor: replace `github.com/golang/mock` with `go.uber.org/mock`

## v1.2.0-alpha1

* feat: apply plan updates
* feat: use CR `meta.name` as default database name
* fix: avoid resource update conflicts during reconcile
* build(deps): update github.com/Scalingo/go-scalingo to v10
* chore(go) update go version to 1.26

## v1.1.0-alpha1

* feat: add and remove firewall rules
* feat: make internet access mandatory

## v1.0.0-alpha1

* feat: deploy/undeploy PostgreSQL instances hosted on dedicated resources
