# Changelog

Changes in this fork relative to upstream [d3vilh/openvpn-ui](https://github.com/d3vilh/openvpn-ui).

## [1.0.0] - 2026-11-06

First release of the fork. Based on upstream `0.9.5.6`.

### Added
- **SMTP email alerts** (Configuration > OpenVPN UI > Email (SMTP) Alerts):
  - SMTP host, port, username/password, from address and name.
  - Implicit TLS (port 465) and STARTTLS (587), with an option to skip
    certificate verification; authentication only when a username is set.
  - Per-event switches: certificate created, certificate revoked, user created.
  - New-certificate emails attach the `.ovpn` profile and, when 2FA is enabled,
    the 2FA QR code (`.png`).
- **Customisable email text**: subject and body for each event can be edited
  under "Email message text", using Go template placeholders such as
  `{{.ClientName}}`. Empty fields keep the built-in message.

### Fixed
- Certificate details popup did not open for names containing a period
  ([d3vilh/openvpn-ui#96](https://github.com/d3vilh/openvpn-ui/issues/96)).
  Modal ids are now sanitised with a new `safeid` template function.
- Renew, Revoke and Delete returned a 404 for certificates with a period in
  the name, or whose `index.txt` entry has no `/name=`, `/LocalIP=` or
  `/2FAName=` suffix
  ([d3vilh/openvpn-ui#97](https://github.com/d3vilh/openvpn-ui/issues/97)).
  Links now fall back to the CN, `none` and `dynamic.pool`, so a URL never
  contains an empty segment.
- Certificate list shows the CN when the name is empty.
- Revoke email lookup matches the certificate by name or CN.
- Default HTML email bodies escape user-supplied values.
- New-user emails show the new user's login instead of the admin's.

### Changed
- Version is now defined in `lib/version.go` and shown in the page footer.
- Repository links in the footer, login page, docs and README point to this fork; the upstream project is credited in the
  footer and README.
- Go module path renamed from `github.com/d3vilh/openvpn-ui` to `github.com/dubey-omprakash/openvpn-ui` (go.mod, imports,
  generated router file, `build/build.sh`). The dependency `d3vilh/openvpn-server-config` is unchanged.
- Added `THIRD-PARTY-NOTICES.md`, the fork's copyright line in `LICENSE`, and a hobby-project/no-warranty notice in the README.
- Neutral placeholder texts in the SMTP settings form.
- Removed the upstream `FUNDING.yml`; the sample `docs/docker-compose.yml` uses locally built images.

### Notes
- The SMTP password is stored as plain text in the UI's SQLite database.
  Restrict access to the `db` volume.
