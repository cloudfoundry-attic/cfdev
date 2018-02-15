package resource

type Catalog struct {
	Items []Item
}

type Item struct {
	URL  string
	Name string
	MD5  string
}