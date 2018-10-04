package ifttt

import "github.com/Jeffail/gabs"

// DynamicOption describes the result of a dynamic field options request.
// To created a nested OptionList, add the inner options as a *DynamicOption and pass it to the outer DynamicOption by (*DynamicOption).AddCategory
type DynamicOption struct {
	options map[string]interface{}
}

// AddCategory adds a category item to the DynamicOption
// Note that the values inside the category could be selected but not the category itself.
func (c *DynamicOption) AddCategory(name string, value *DynamicOption) {
	if c.options == nil {
		c.options = make(map[string]interface{})
	}
	c.options[name] = value
}

// AddString adds an option value labeled by name to the DynamicOption
func (c *DynamicOption) AddString(name string, value string) {
	if c.options == nil {
		c.options = make(map[string]interface{})
	}
	c.options[name] = value
}

func (c *DynamicOption) packThis() *gabs.Container {
	obj := gabs.New()
	obj.Array()
	for key, val := range c.options {
		this := gabs.New()
		this.Set(key, "label")
		switch val.(type) {
		case string:
			this.Set(val, "value")
		case *DynamicOption:
			this.Set(val.(*DynamicOption).packThis().Data(), "values")
		}
		obj.ArrayAppend(this.Data())
	}
	return obj
}

func (c *DynamicOption) marshal() []byte {
	res := gabs.New()

	res.Set(c.packThis().Data(), "data")
	return res.Bytes()
}
