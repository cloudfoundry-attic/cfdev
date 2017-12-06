package resource

type Catalog struct {
	Items []Item
}

type Item struct {
	URL  string
	Name string
	MD5  string

	OS string // The operating system this item targets
}

func (c *Catalog) Filter(os string) *Catalog {
	var copy Catalog

	for _, item := range c.Items {
		if item.OS == "" || item.OS == os {
			copy.Items = append(copy.Items, item)
		}
	}

	return &copy
}
