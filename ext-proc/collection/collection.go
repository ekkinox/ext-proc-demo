package collection

import (
	"fmt"
)

type Collection struct {
	Items map[string]string
}

func (c *Collection) All() map[string]string {
	return c.Items
}

func (c *Collection) Count() int {
	return len(c.Items)
}

func (c *Collection) Has(key string) bool {

	if _, ok := c.Items[key]; ok {
		return true
	}

	return false
}

func (c *Collection) Set(key string, value string) *Collection {

	c.Items[key] = value

	return c
}

func (c *Collection) Get(key string) (string, error) {

	if c.Has(key) {
		return c.Items[key], nil
	} else {
		return "", fmt.Errorf("could not find key %s in collection", key)
	}
}
