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
	ID         string
	CategoryID string
	Name       string
	Quantity   string
	InBag      bool // The bag denotes in-progress item additions
	Done       bool
}
