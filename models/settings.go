package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
	//Sqlite driver
	_ "github.com/mattn/go-sqlite3"
)

type Settings struct {
	Id      int64
	Profile string `orm:"size(64);unique" form:"Profile" valid:"Required;"`

	MIAddress string `orm:"size(64);unique" form:"MIAddress" valid:"Required;"`
	MINetwork string `orm:"size(64);unique" form:"MINetwork" valid:"Required;"`

	OVConfigPath string `orm:"size(64);unique" form:"OVConfigPath" valid:"Required;"`
	EasyRSAPath  string `orm:"size(64);unique" form:"EasyRSAPath" valid:"Required;"`

	//	ServerAddress     string `orm:"size(64);unique" form:"ServerAddress" valid:"Required;"`
	//	OpenVpnServerPort string `orm:"size(64);unique" form:"OpenVpnServerPort" valid:"Required;"`

	// SMTP / Email alert configuration
	SMTPEnabled    bool   `orm:"default(false)" form:"SMTPEnabled"`
	SMTPHost       string `orm:"size(128);null" form:"SMTPHost"`
	SMTPPort       int    `orm:"default(587);null" form:"SMTPPort"`
	SMTPUsername   string `orm:"size(128);null" form:"SMTPUsername"`
	SMTPPassword   string `orm:"size(128);null" form:"SMTPPassword"`
	SMTPFromEmail  string `orm:"size(128);null" form:"SMTPFromEmail"`
	SMTPFromName   string `orm:"size(128);null" form:"SMTPFromName"`
	SMTPUseTLS     bool   `orm:"default(true)" form:"SMTPUseTLS"`
	SMTPSkipVerify bool   `orm:"default(false)" form:"SMTPSkipVerify"`
	// Which events trigger an outbound email
	SMTPNotifyOnCertCreate bool `orm:"default(true)" form:"SMTPNotifyOnCertCreate"`
	SMTPNotifyOnCertRevoke bool `orm:"default(true)" form:"SMTPNotifyOnCertRevoke"`
	SMTPNotifyOnUserCreate bool `orm:"default(true)" form:"SMTPNotifyOnUserCreate"`

	// Optional custom email text. Empty means "use the built-in message".
	// Subjects and bodies are Go text/template strings, see lib/email.go for the
	// available placeholders (.SiteName, .ClientName, .Serial, ...).
	EmailSiteName      string `orm:"size(128);null" form:"EmailSiteName"`
	EmailCertSubject   string `orm:"size(255);null" form:"EmailCertSubject"`
	EmailCertBody      string `orm:"type(text);null" form:"EmailCertBody"`
	EmailRevokeSubject string `orm:"size(255);null" form:"EmailRevokeSubject"`
	EmailRevokeBody    string `orm:"type(text);null" form:"EmailRevokeBody"`
	EmailUserSubject   string `orm:"size(255);null" form:"EmailUserSubject"`
	EmailUserBody      string `orm:"type(text);null" form:"EmailUserBody"`

	Created time.Time `orm:"auto_now_add;type(datetime)"`
	Updated time.Time `orm:"auto_now;type(datetime)"`
}

// Insert wrapper
func (s *Settings) Insert() error {
	if _, err := orm.NewOrm().Insert(s); err != nil {
		return err
	}
	return nil
}

// Read wrapper
func (s *Settings) Read(fields ...string) error {
	if err := orm.NewOrm().Read(s, fields...); err != nil {
		return err
	}
	return nil
}

// Update wrapper
func (s *Settings) Update(fields ...string) error {
	if _, err := orm.NewOrm().Update(s, fields...); err != nil {
		return err
	}
	return nil
}

// Delete wrapper
func (s *Settings) Delete() error {
	if _, err := orm.NewOrm().Delete(s); err != nil {
		return err
	}
	return nil
}
