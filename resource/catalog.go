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

func (c *Catalog) Lookup(name string) *Item {
	for index := range c.Items {
		item := &c.Items[index]
		if item.Name == name {
			return item
		}
	}
	return nil
}

func (c *Catalog) Remove(name string) {
	newItems := make([]Item, 0, len(c.Items))
	for _, item := range c.Items {
		if item.Name != name {
			newItems = append(newItems, item)
		}
	}
	c.Items = newItems
}
