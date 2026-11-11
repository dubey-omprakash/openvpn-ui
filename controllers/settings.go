package controllers

import (
	"html/template"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/dubey-omprakash/openvpn-ui/models"
	"github.com/dubey-omprakash/openvpn-ui/state"
)

type SettingsController struct {
	BaseController
}

func (c *SettingsController) NestPrepare() {
	if !c.IsLogin {
		c.Ctx.Redirect(302, c.LoginPath())
		return
	}
	c.Data["breadcrumbs"] = &BreadCrumbs{
		Title: "OpenVPN UI Settings",
	}
}

func (c *SettingsController) Get() {
	c.TplName = "settings.html"
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML())
	settings := models.Settings{Profile: "default"}
	_ = settings.Read("Profile")
	c.Data["Settings"] = &settings
}

func (c *SettingsController) Post() {
	c.TplName = "settings.html"

	flash := web.NewFlash()
	settings := models.Settings{Profile: "default"}
	_ = settings.Read("Profile")
	if err := c.ParseForm(&settings); err != nil {
		logs.Warning(err)
		flash.Error(err.Error())
		flash.Store(&c.Controller)
		return
	}

	// Checkbox fields are omitted entirely from the POST body when unchecked,
	// so ParseForm cannot distinguish "left unchecked" from "field absent" --
	// set them explicitly from the raw form values instead.
	settings.SMTPEnabled = c.GetString("SMTPEnabled") == "on"
	settings.SMTPUseTLS = c.GetString("SMTPUseTLS") == "on"
	settings.SMTPSkipVerify = c.GetString("SMTPSkipVerify") == "on"
	settings.SMTPNotifyOnCertCreate = c.GetString("SMTPNotifyOnCertCreate") == "on"
	settings.SMTPNotifyOnCertRevoke = c.GetString("SMTPNotifyOnCertRevoke") == "on"
	settings.SMTPNotifyOnUserCreate = c.GetString("SMTPNotifyOnUserCreate") == "on"

	// If the password field was left blank (e.g. the operator only changed
	// the host), keep the previously stored password instead of wiping it.
	if c.GetString("SMTPPassword") == "" {
		existing := models.Settings{Profile: "default"}
		if err := existing.Read("Profile"); err == nil {
			settings.SMTPPassword = existing.SMTPPassword
		}
	}

	c.Data["Settings"] = &settings

	o := orm.NewOrm()
	if _, err := o.Update(&settings); err != nil {
		flash.Error(err.Error())
	} else {
		flash.Success("Settings has been updated")
		state.GlobalCfg = settings
	}
	flash.Store(&c.Controller)
}
