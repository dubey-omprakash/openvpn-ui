package lib

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"html"
	"mime"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/d3vilh/openvpn-ui/models"
)

// EmailAttachment represents a single file to attach to an outbound message.
type EmailAttachment struct {
	Filename string
	Data     []byte
	MimeType string // e.g. "application/x-openvpn-profile", "image/png"
}

// EmailMessage is the generic message envelope used by SendEmail.
type EmailMessage struct {
	To          []string
	Subject     string
	BodyText    string
	BodyHTML    string
	Attachments []EmailAttachment
}

// loadSMTPSettings reads the current "default" settings profile from the DB.
func loadSMTPSettings() (*models.Settings, error) {
	settings := &models.Settings{Profile: "default"}
	if err := settings.Read("Profile"); err != nil {
		return nil, fmt.Errorf("unable to read settings: %w", err)
	}
	return settings, nil
}

// SendEmail sends an arbitrary message (with optional attachments) using the
// SMTP configuration stored in Settings. It is safe to call even when SMTP is
// disabled or unconfigured -- in that case it logs and returns nil so callers
// (certificate/user hooks) never fail the parent operation because email
// could not be sent.
func SendEmail(msg EmailMessage) error {
	settings, err := loadSMTPSettings()
	if err != nil {
		logs.Error("Email: unable to load settings: %v", err)
		return err
	}

	if !settings.SMTPEnabled {
		logs.Debug("Email: SMTP notifications are disabled, skipping send to %v", msg.To)
		return nil
	}

	if settings.SMTPHost == "" || settings.SMTPPort == 0 {
		return fmt.Errorf("SMTP host/port is not configured")
	}

	if len(msg.To) == 0 {
		return fmt.Errorf("no recipients specified for email %q", msg.Subject)
	}

	fromAddr := settings.SMTPFromEmail
	if fromAddr == "" {
		fromAddr = settings.SMTPUsername
	}
	fromName := settings.SMTPFromName
	if fromName == "" {
		fromName = "OpenVPN UI"
	}

	raw, err := buildMIMEMessage(fromName, fromAddr, msg)
	if err != nil {
		return fmt.Errorf("failed to build MIME message: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", settings.SMTPHost, settings.SMTPPort)

	var auth smtp.Auth
	if settings.SMTPUsername != "" {
		auth = smtp.PlainAuth("", settings.SMTPUsername, settings.SMTPPassword, settings.SMTPHost)
	}

	sendErr := sendViaSMTP(addr, settings.SMTPHost, auth, fromAddr, msg.To, raw, settings.SMTPUseTLS, settings.SMTPSkipVerify)
	if sendErr != nil {
		logs.Error("Email: failed to send %q to %v: %v", msg.Subject, msg.To, sendErr)
		return sendErr
	}

	logs.Info("Email: sent %q to %v", msg.Subject, msg.To)
	return nil
}

// sendViaSMTP dials the SMTP server and transmits the message. It supports
// both implicit TLS (port 465 style) and STARTTLS (port 587/25 style),
// falling back to plain submission only if SMTPUseTLS is explicitly false --
// which is not recommended for production/tender deployments.
func sendViaSMTP(addr, host string, auth smtp.Auth, from string, to []string, msg []byte, useTLS bool, skipVerify bool) error {
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: skipVerify, // #nosec G402 - operator opt-in only, surfaced in UI
	}

	// Implicit TLS on the classic SMTPS port.
	if useTLS && strings.HasSuffix(addr, ":465") {
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("tls dial failed: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("smtp client init failed: %w", err)
		}
		defer client.Close()

		return submit(client, auth, from, to, msg)
	}

	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial failed: %w", err)
	}
	defer client.Close()

	if useTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("starttls failed: %w", err)
			}
		} else {
			logs.Warn("Email: server %s does not advertise STARTTLS, sending unencrypted", host)
		}
	}

	return submit(client, auth, from, to, msg)
}

func submit(client *smtp.Client, auth smtp.Auth, from string, to []string, msg []byte) error {
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth failed: %w", err)
		}
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("RCPT TO %s failed: %w", addr, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return fmt.Errorf("writing message body failed: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("closing DATA writer failed: %w", err)
	}
	return client.Quit()
}

// buildMIMEMessage assembles a multipart/mixed RFC 5322 message with an
// optional multipart/alternative text+HTML body and any number of
// base64-encoded attachments (used for the .ovpn profile and 2FA QR PNG).
func buildMIMEMessage(fromName, fromAddr string, msg EmailMessage) ([]byte, error) {
	var buf bytes.Buffer

	boundaryMixed := fmt.Sprintf("mixed-%d", time.Now().UnixNano())

	buf.WriteString(fmt.Sprintf("From: %s <%s>\r\n", mime.QEncoding.Encode("UTF-8", fromName), fromAddr))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", msg.Subject)))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	buf.WriteString("MIME-Version: 1.0\r\n")

	if len(msg.Attachments) == 0 {
		writeBodyPart(&buf, msg, "")
	} else {
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundaryMixed))
		writeBodyPart(&buf, msg, boundaryMixed)

		for _, att := range msg.Attachments {
			mimeType := att.MimeType
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundaryMixed))
			buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", mimeType, att.Filename))
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", att.Filename))
			buf.WriteString(chunkBase64(att.Data))
			buf.WriteString("\r\n")
		}
		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundaryMixed))
	}

	return buf.Bytes(), nil
}

func writeBodyPart(buf *bytes.Buffer, msg EmailMessage, outerBoundary string) {
	if outerBoundary != "" {
		buf.WriteString(fmt.Sprintf("--%s\r\n", outerBoundary))
	}

	if msg.BodyHTML != "" && msg.BodyText != "" {
		boundaryAlt := fmt.Sprintf("alt-%d", time.Now().UnixNano())
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundaryAlt))

		buf.WriteString(fmt.Sprintf("--%s\r\n", boundaryAlt))
		buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
		buf.WriteString(msg.BodyText + "\r\n\r\n")

		buf.WriteString(fmt.Sprintf("--%s\r\n", boundaryAlt))
		buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
		buf.WriteString(msg.BodyHTML + "\r\n\r\n")

		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundaryAlt))
	} else if msg.BodyHTML != "" {
		buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
		buf.WriteString(msg.BodyHTML + "\r\n")
	} else {
		buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
		buf.WriteString(msg.BodyText + "\r\n")
	}

	if outerBoundary == "" {
		return
	}
	buf.WriteString("\r\n")
}

func chunkBase64(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	var out strings.Builder
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		out.WriteString(encoded[i:end])
		out.WriteString("\r\n")
	}
	return out.String()
}

// AlertKind enumerates the events SendAlertEmail knows how to render.
type AlertKind string

const (
	AlertCertificateCreated AlertKind = "certificate_created"
	AlertCertificateRevoked AlertKind = "certificate_revoked"
	AlertUserCreated        AlertKind = "user_created"
)

// AlertData carries everything a given alert template might need. Not every
// field is required for every AlertKind.
type AlertData struct {
	RecipientEmail string
	RecipientName  string
	ClientName     string // certificate / VPN client CN
	Serial         string
	StaticIP       string
	ExpireDays     string
	ActorLogin     string // admin/user who performed the action
	Login          string // login name of a newly created UI user
	SiteName       string
	OVPNProfile    []byte // raw .ovpn file content, attached when present
	OVPNFilename   string
	QRCodePNG      []byte // raw QR PNG content, attached when present (2FA certs)
	QRFilename     string
}

// SendAlertEmail renders a template for the given AlertKind and dispatches it
// via SendEmail, attaching the .ovpn profile and/or 2FA QR code when
// provided. Errors are logged and returned but are intentionally
// non-fatal to the caller's primary workflow (certificate/user creation
// must never fail purely because the alert email could not be sent).
func SendAlertEmail(kind AlertKind, data AlertData) error {
	settings, err := loadSMTPSettings()
	if err != nil {
		logs.Error("SendAlertEmail: unable to load settings: %v", err)
		return err
	}
	if !settings.SMTPEnabled {
		return nil
	}

	switch kind {
	case AlertCertificateCreated:
		if !settings.SMTPNotifyOnCertCreate {
			return nil
		}
	case AlertCertificateRevoked:
		if !settings.SMTPNotifyOnCertRevoke {
			return nil
		}
	case AlertUserCreated:
		if !settings.SMTPNotifyOnUserCreate {
			return nil
		}
	}

	if data.RecipientEmail == "" {
		logs.Warn("SendAlertEmail: no recipient email for %s / %s, skipping", kind, data.ClientName)
		return nil
	}

	siteName := data.SiteName
	if siteName == "" {
		siteName = strings.TrimSpace(settings.EmailSiteName)
	}
	if siteName == "" {
		siteName = "OpenVPN UI"
	}

	subject, bodyText, bodyHTML := renderAlertTemplate(settings, kind, data, siteName)

	msg := EmailMessage{
		To:       []string{data.RecipientEmail},
		Subject:  subject,
		BodyText: bodyText,
		BodyHTML: bodyHTML,
	}

	if len(data.OVPNProfile) > 0 {
		filename := data.OVPNFilename
		if filename == "" {
			filename = sanitizeFilename(data.ClientName) + ".ovpn"
		}
		msg.Attachments = append(msg.Attachments, EmailAttachment{
			Filename: filename,
			Data:     data.OVPNProfile,
			MimeType: "application/x-openvpn-profile",
		})
	}

	if len(data.QRCodePNG) > 0 {
		filename := data.QRFilename
		if filename == "" {
			filename = sanitizeFilename(data.ClientName) + "-2fa-qr.png"
		}
		msg.Attachments = append(msg.Attachments, EmailAttachment{
			Filename: filename,
			Data:     data.QRCodePNG,
			MimeType: "image/png",
		})
	}

	return SendEmail(msg)
}

func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", " ", "_", "\"", "")
	return replacer.Replace(name)
}

func renderAlertTemplate(settings *models.Settings, kind AlertKind, data AlertData, siteName string) (subject, bodyText, bodyHTML string) {
	// Values placed into the default HTML bodies must be escaped.
	eh := data
	eh.ClientName = html.EscapeString(data.ClientName)
	eh.Serial = html.EscapeString(data.Serial)
	eh.StaticIP = html.EscapeString(data.StaticIP)
	eh.ExpireDays = html.EscapeString(data.ExpireDays)
	eh.ActorLogin = html.EscapeString(data.ActorLogin)
	eh.RecipientName = html.EscapeString(data.RecipientName)
	eh.Login = html.EscapeString(data.Login)
	esite := html.EscapeString(siteName)

	switch kind {
	case AlertCertificateCreated:
		subject = fmt.Sprintf("[%s] New VPN certificate issued: %s", siteName, data.ClientName)
		bodyText = fmt.Sprintf(
			"Hello,\n\nA new VPN certificate has been issued for \"%s\".\n\n"+
				"Static IP: %s\nCertificate validity: %s day(s)\nIssued by: %s\n\n"+
				"Your .ovpn client profile is attached. If two-factor authentication\n"+
				"is enabled, the QR code for your authenticator app is attached as well.\n\n"+
				"Keep these files confidential -- they grant access to the VPN.\n\n-- %s\n",
			data.ClientName, orDash(data.StaticIP), orDash(data.ExpireDays), orDash(data.ActorLogin), siteName)
		bodyHTML = fmt.Sprintf(
			"<p>Hello,</p><p>A new VPN certificate has been issued for <b>%s</b>.</p>"+
				"<ul><li>Static IP: %s</li><li>Certificate validity: %s day(s)</li><li>Issued by: %s</li></ul>"+
				"<p>Your <b>.ovpn</b> client profile is attached. If two-factor authentication is enabled, "+
				"the QR code for your authenticator app is attached as well.</p>"+
				"<p style=\"color:#b00\">Keep these files confidential — they grant access to the VPN.</p>"+
				"<p>-- %s</p>",
			eh.ClientName, orDash(eh.StaticIP), orDash(eh.ExpireDays), orDash(eh.ActorLogin), esite)

	case AlertCertificateRevoked:
		subject = fmt.Sprintf("[%s] VPN certificate revoked: %s", siteName, data.ClientName)
		bodyText = fmt.Sprintf(
			"Hello,\n\nThe VPN certificate for \"%s\" (serial %s) has been revoked by %s.\n"+
				"This client will no longer be able to connect once the server reloads its CRL.\n\n-- %s\n",
			data.ClientName, orDash(data.Serial), orDash(data.ActorLogin), siteName)
		bodyHTML = fmt.Sprintf(
			"<p>Hello,</p><p>The VPN certificate for <b>%s</b> (serial %s) has been revoked by %s.</p>"+
				"<p>This client will no longer be able to connect once the server reloads its CRL.</p><p>-- %s</p>",
			eh.ClientName, orDash(eh.Serial), orDash(eh.ActorLogin), esite)

	case AlertUserCreated:
		subject = fmt.Sprintf("[%s] Your OpenVPN UI account was created", siteName)
		bodyText = fmt.Sprintf(
			"Hello %s,\n\nAn OpenVPN UI administration account has been created for you on %s.\n"+
				"Login: %s\n\nPlease sign in and change your password on first use.\n\n-- %s\n",
			orDash(data.RecipientName), siteName, orDash(data.Login), siteName)
		bodyHTML = fmt.Sprintf(
			"<p>Hello %s,</p><p>An OpenVPN UI administration account has been created for you on <b>%s</b>.</p>"+
				"<p>Login: <b>%s</b></p><p>Please sign in and change your password on first use.</p><p>-- %s</p>",
			orDash(eh.RecipientName), esite, orDash(eh.Login), esite)

	default:
		subject = fmt.Sprintf("[%s] Notification", siteName)
		bodyText = "This is an automated notification."
		bodyHTML = "<p>This is an automated notification.</p>"
	}
	return applyCustomTemplates(settings, kind, data, siteName, subject, bodyText, bodyHTML)
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

// readFileIfExists is a small helper used by controller hooks to attach the
// .ovpn profile / QR PNG without failing the whole request if the file has
// not been generated yet (e.g. static clients, or 2FA disabled).
func readFileIfExists(path string) []byte {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			logs.Warn("Email: could not read attachment %s: %v", path, err)
		}
		return nil
	}
	return data
}

// ReadOVPNProfile reads a generated .ovpn client profile from the clients
// directory for the given OpenVPN config path and client name.
func ReadOVPNProfile(ovConfigPath, name string) []byte {
	return readFileIfExists(filepath.Join(ovConfigPath, "clients", name+".ovpn"))
}

// ReadQRCode reads a generated 2FA QR PNG from the clients directory for the
// given OpenVPN config path and client name.
func ReadQRCode(ovConfigPath, name string) []byte {
	return readFileIfExists(filepath.Join(ovConfigPath, "clients", name+".png"))
}

// templateVars are the placeholders an administrator can use in custom email
// subjects and bodies, for example: "Certificate {{.ClientName}} was issued".
type templateVars struct {
	SiteName       string
	ClientName     string
	Serial         string
	StaticIP       string
	ExpireDays     string
	ActorLogin     string
	RecipientName  string
	RecipientEmail string
	Login          string
}

func renderCustomText(name, tpl string, v templateVars) (string, error) {
	t, err := template.New(name).Parse(tpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, v); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// textToHTML converts a plain-text body to a simple, escaped HTML body.
func textToHTML(s string) string {
	escaped := html.EscapeString(strings.ReplaceAll(s, "\r\n", "\n"))
	return "<div style=\"font-family:sans-serif\">" + strings.ReplaceAll(escaped, "\n", "<br>\n") + "</div>"
}

// applyCustomTemplates replaces the built-in subject/body with the
// administrator's own text from Settings, when one is configured. A template
// that fails to parse or run is logged and the built-in text is used instead,
// so a typo in a template never stops an alert from being sent.
func applyCustomTemplates(settings *models.Settings, kind AlertKind, data AlertData, siteName, subject, bodyText, bodyHTML string) (string, string, string) {
	var subjTpl, bodyTpl string
	switch kind {
	case AlertCertificateCreated:
		subjTpl, bodyTpl = settings.EmailCertSubject, settings.EmailCertBody
	case AlertCertificateRevoked:
		subjTpl, bodyTpl = settings.EmailRevokeSubject, settings.EmailRevokeBody
	case AlertUserCreated:
		subjTpl, bodyTpl = settings.EmailUserSubject, settings.EmailUserBody
	}

	vars := templateVars{
		SiteName:       siteName,
		ClientName:     data.ClientName,
		Serial:         data.Serial,
		StaticIP:       orDash(data.StaticIP),
		ExpireDays:     orDash(data.ExpireDays),
		ActorLogin:     data.ActorLogin,
		RecipientName:  data.RecipientName,
		RecipientEmail: data.RecipientEmail,
		Login:          data.Login,
	}

	if strings.TrimSpace(subjTpl) != "" {
		if out, err := renderCustomText("subject", subjTpl, vars); err != nil {
			logs.Warn("Email: custom subject for %s is invalid, using the default: %v", kind, err)
		} else {
			// A subject must be a single line.
			out = strings.NewReplacer("\r", " ", "\n", " ").Replace(out)
			if strings.TrimSpace(out) != "" {
				subject = strings.TrimSpace(out)
			}
		}
	}

	if strings.TrimSpace(bodyTpl) != "" {
		if out, err := renderCustomText("body", bodyTpl, vars); err != nil {
			logs.Warn("Email: custom body for %s is invalid, using the default: %v", kind, err)
		} else if strings.TrimSpace(out) != "" {
			bodyText = out
			bodyHTML = textToHTML(out)
		}
	}

	return subject, bodyText, bodyHTML
}
