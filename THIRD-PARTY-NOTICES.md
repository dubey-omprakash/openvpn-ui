# Third-party notices

This project is a personal hobby project released under the MIT licence (see [LICENSE](LICENSE)). It builds on the software
below. Each component keeps its own licence; the full licence texts are in the `vendor/<module>/` folders of this repository
(Go modules) or in the file headers (bundled front-end files). This list was generated from `vendor/modules.txt` and the
licence files found there; licence names are detected automatically, so check the licence file if it matters to you.

## Original project

| Project | Licence | Notes |
|---|---|---|
| [d3vilh/openvpn-ui](https://github.com/d3vilh/openvpn-ui) | MIT | This repository is a fork. Copyright (c) 2017-2021 Adam Wałach, (c) 2021 Mr.Philipp aka d3vilh. |
| [d3vilh/openvpn-server-config](https://github.com/d3vilh/openvpn-server-config) | MIT | Go library used as a dependency (not forked). |

## Go modules (vendored)

| Module | Version | Licence |
|---|---|---|
| `cloud.google.com/go/auth` | v0.22.0 | Apache-2.0 |
| `cloud.google.com/go/auth/oauth2adapt` | v0.2.8 | Apache-2.0 |
| `cloud.google.com/go/compute/metadata` | v0.9.0 | Apache-2.0 |
| `github.com/Azure/go-ntlmssp` | v0.1.1 | MIT |
| `github.com/beego/beego/v2` | v2.3.10 | Apache-2.0 |
| `github.com/beorn7/perks` | v1.0.1 | MIT |
| `github.com/cespare/xxhash/v2` | v2.3.0 | MIT |
| `github.com/cloudfoundry/gosigar` | v1.3.125 | Apache-2.0 |
| `github.com/d3vilh/openvpn-server-config` | v0.3.5 | Apache-2.0 |
| `github.com/felixge/httpsnoop` | v1.0.4 | MIT |
| `github.com/go-asn1-ber/asn1-ber` | v1.5.8 | MIT |
| `github.com/go-ldap/ldap/v3` | v3.4.14 | MIT |
| `github.com/go-logr/logr` | v1.4.3 | Apache-2.0 |
| `github.com/go-logr/stdr` | v1.2.2 | Apache-2.0 |
| `github.com/google/s2a-go` | v0.1.9 | Apache-2.0 |
| `github.com/google/uuid` | v1.6.0 | BSD-3-Clause |
| `github.com/googleapis/enterprise-certificate-proxy` | v0.3.19 | Apache-2.0 |
| `github.com/googleapis/gax-go/v2` | v2.23.0 | BSD-3-Clause |
| `github.com/hashicorp/golang-lru` | v1.0.2 | MPL-2.0 |
| `github.com/mattn/go-sqlite3` | v1.14.49 | MIT |
| `github.com/mitchellh/mapstructure` | v1.5.0 | MIT |
| `github.com/munnerz/goautoneg` | v0.0.0-20191010083416-a7dc8b61c822 | BSD-3-Clause |
| `github.com/prometheus/client_golang` | v1.22.0 | Apache-2.0 |
| `github.com/prometheus/client_model` | v0.6.2 | Apache-2.0 |
| `github.com/prometheus/common` | v0.63.0 | Apache-2.0 |
| `github.com/prometheus/procfs` | v0.16.1 | Apache-2.0 |
| `github.com/shiena/ansicolor` | v0.0.0-20230509054315-a9deabde6e02 | MIT |
| `github.com/sirupsen/logrus` | v1.9.4 | MIT |
| `github.com/valyala/bytebufferpool` | v1.0.0 | MIT |
| `go.opentelemetry.io/auto/sdk` | v1.2.1 | Apache-2.0 |
| `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` | v0.67.0 | Apache-2.0 |
| `go.opentelemetry.io/otel` | v1.44.0 | Apache-2.0 |
| `go.opentelemetry.io/otel/metric` | v1.44.0 | Apache-2.0 |
| `go.opentelemetry.io/otel/trace` | v1.44.0 | Apache-2.0 |
| `golang.org/x/crypto` | v0.54.0 | BSD-3-Clause |
| `golang.org/x/net` | v0.57.0 | BSD-3-Clause |
| `golang.org/x/oauth2` | v0.36.0 | BSD-3-Clause |
| `golang.org/x/sys` | v0.47.0 | BSD-3-Clause |
| `golang.org/x/text` | v0.40.0 | BSD-3-Clause |
| `google.golang.org/api` | v0.291.0 | BSD-3-Clause |
| `google.golang.org/genproto/googleapis/rpc` | v0.0.0-20260724162435-b2f20204f0df | Apache-2.0 |
| `google.golang.org/grpc` | v1.82.1 | Apache-2.0 |
| `google.golang.org/protobuf` | v1.36.11 | BSD-3-Clause |
| `gopkg.in/hlandau/passlib.v1` | v1.0.11 | BSD-3-Clause |
| `gopkg.in/yaml.v3` | v3.0.1 | Apache-2.0 |
| `gopkg.in/hlandau/easymetric.v1` | v1.0.0 | no licence file vendored; see the upstream project |
| `gopkg.in/hlandau/measurable.v1` | v1.0.1 | no licence file vendored; see the upstream project |

## Front-end files bundled in `static/`

| Component | Licence |
|---|---|
| [AdminLTE](https://adminlte.io) 3.2.0 | MIT |
| [Bootstrap](https://getbootstrap.com) 4.6.1 | MIT |
| [jQuery](https://jquery.com) 3.6.1 | MIT |
| [clipboard.js](https://clipboardjs.com) 2.0.11 | MIT |
| [iCheck](https://github.com/fronteed/iCheck) 1.0.3 | MIT |
| [Ace editor](https://ace.c9.io) (`static/js/ace`) | BSD-3-Clause |
| Font Awesome 4, Glyphicons, Ionicons (`static/fonts`) | see the respective projects (SIL OFL 1.1 / MIT) |

## Software used at runtime or build time (not included in this repository)

| Software | Licence |
|---|---|
| [OpenVPN](https://openvpn.net) | GPL-2.0 (with OpenSSL exception) |
| [EasyRSA](https://github.com/OpenVPN/easy-rsa) | GPL-2.0 |
| [Alpine Linux](https://alpinelinux.org) base image | various open source licences |
| [qrencode](https://github.com/d3vilh/qrencode) (built into the UI image; uses libqrencode, LGPL-2.1) | see upstream |

"OpenVPN" is a registered trademark of OpenVPN Inc. This project is not affiliated with, endorsed by or sponsored by OpenVPN Inc.
