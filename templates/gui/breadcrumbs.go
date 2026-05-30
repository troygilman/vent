package gui

import "fmt"

type BreadcrumbItem struct {
	Label string
	Href  string // empty means current page
}

func HomeBreadcrumbs() []BreadcrumbItem {
	return []BreadcrumbItem{
		{Label: "Home"},
	}
}

func SchemaListBreadcrumbs(pluralName string) []BreadcrumbItem {
	return []BreadcrumbItem{
		{Label: pluralName},
	}
}

func SchemaAddBreadcrumbs(adminPath, routeName, pluralName, singularName string) []BreadcrumbItem {
	listPath := adminPath + routeName + "/"
	return []BreadcrumbItem{
		{Label: pluralName, Href: listPath},
		{Label: "Add " + singularName},
	}
}

func SchemaEntityBreadcrumbs(adminPath, routeName, pluralName, entityDisplay string) []BreadcrumbItem {
	listPath := adminPath + routeName + "/"
	return []BreadcrumbItem{
		{Label: pluralName, Href: listPath},
		{Label: entityDisplay},
	}
}

func SchemaPasswordBreadcrumbs(adminPath, routeName, pluralName, entityDisplay string, entityID int) []BreadcrumbItem {
	listPath := adminPath + routeName + "/"
	entityPath := fmt.Sprintf("%s%s/%d/", adminPath, routeName, entityID)
	return []BreadcrumbItem{
		{Label: pluralName, Href: listPath},
		{Label: entityDisplay, Href: entityPath},
		{Label: "Manage Password"},
	}
}
