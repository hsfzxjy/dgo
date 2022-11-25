package dgo

type InvokeContext struct {
	flag CallbackFlag
	port *Port
}

func (c *InvokeContext) Flag() CallbackFlag { return c.flag }
func (c *InvokeContext) Port() *Port        { return c.port }
