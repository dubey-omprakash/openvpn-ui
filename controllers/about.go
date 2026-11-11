package controllers

import (
	"os"

	"github.com/dubey-omprakash/openvpn-ui/lib"
)

// AboutController shows version, credits and licence information.
type AboutController struct {
	BaseController
}

func (c *AboutController) NestPrepare() {
	if !c.IsLogin {
		c.Ctx.Redirect(302, c.LoginPath())
		return
	}
}

func (c *AboutController) Get() {
	c.TplName = "about.html"
	c.Data["breadcrumbs"] = &BreadCrumbs{
		Title: "About",
	}
	c.Data["AppVersion"] = lib.AppVersion
	// The licence text is shipped with the application; show it when it can be read.
	if text, err := os.ReadFile("LICENSE"); err == nil {
		c.Data["LicenseText"] = string(text)
	}
}
