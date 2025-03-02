package models

const ListDBKey = "list"

type List struct {
	Categories []*Category
}

type Category struct {
	ID    string
	Name  string
	Items []Item
}

type Item struct {
	ID       string
	Category string
	Name     string
	Quantity string
	Done     bool
}
