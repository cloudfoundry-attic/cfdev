package resource

type Catalog struct {
	Items []Item
}

type Item struct {
	URL  string
	Name string
	MD5  string
	Size uint64
	Type string
	InUse bool
}

func (c Catalog) Lookup(name string) *Item {
	for _, item := range c.Items {
		if item.Name == name {
			return &item
		}
	}
	return nil
}