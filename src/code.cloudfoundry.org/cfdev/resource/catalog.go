package resource

type Catalog struct {
	Items []Item
}

type Item struct {
	URL   string
	Name  string
	MD5   string
	Size  uint64
	InUse bool
}

func (c Catalog) Lookup(name string) *Item {
	for index := range c.Items {
		item := &c.Items[index]
		if item.Name == name {
			return item
		}
	}
	return nil
}
