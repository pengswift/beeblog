package controllers

import (
	"beeblog/models"
	"github.com/astaxie/beego"
)

type CategoryController struct {
	beego.Controller
}

func (this *CategoryController) Get() {
	op := this.Input().Get("op")

	switch op {
	case "add":
		name := this.Input().Get("name")
		if len(name) == 0 {
			break
		}
		err := models.AddCategory(name)
		if err != nil {
			beego.Error(err)
		}
		this.Redirect("/category", 301)
		return
	case "del":
		println("ssssssss....................")

		id := this.Input().Get("id")
		if len(id) == 0 {
			break
		}

		println(".......category 1")

		err := models.DelCategory(id)

		println(".......category 2")

		if err != nil {
			beego.Error(err)
		}
		println(".......category 3")
		this.Redirect("/category", 301)
		return
	}

	this.TplNames = "category.html"
	this.Data["IsCategory"] = true

	var err error
	this.Data["Categories"], err = models.GetAllCategories()

	if err != nil {
		beego.Error(err)
	}
}
