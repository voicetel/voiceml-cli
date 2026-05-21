package commands

func registerDiagnostics(r *Registry) {
	g := &Group{Name: "diagnostics", Description: "Health checks and OpenAPI spec (unauthenticated)."}
	g.Commands = []*Command{
		{
			Names:    []string{"health"},
			Synopsis: "Fetch /health status.",
			Usage:    "diagnostics health",
			Run: func(c *Context, _ []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				out, err := c.Client.Diagnostics().Health(c.Ctx)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"openapi"},
			Synopsis: "Fetch /openapi.json.",
			Usage:    "diagnostics openapi",
			Run: func(c *Context, _ []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				out, err := c.Client.Diagnostics().OpenAPI(c.Ctx)
				if err != nil {
					return err
				}
				return c.Printer.JSON(out)
			},
		},
	}
	r.AddGroup(g)
}
