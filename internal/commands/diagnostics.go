package commands

import "github.com/voicetel/voiceml-cli/internal/sdkclient"

func registerDiagnostics(r *Registry) {
	g := &Group{Name: "diagnostics", Description: "Health checks and OpenAPI spec (unauthenticated)."}
	g.Commands = []*Command{
		{
			Names:    []string{"health"},
			Synopsis: "Fetch /health status.",
			Usage:    "diagnostics health",
			Run: func(c *Context, _ []string, _ string) error {
				if err := sdkclient.DiagnosticsError(c.Client); err != nil {
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
				if err := sdkclient.DiagnosticsError(c.Client); err != nil {
					return err
				}
				out, err := c.Client.Diagnostics().OpenAPI(c.Ctx)
				if err != nil {
					return err
				}
				return c.Printer.JSON(out)
			},
		},
		{
			Names:    []string{"openapi-yaml"},
			Synopsis: "Fetch /openapi.yaml.",
			Usage:    "diagnostics openapi-yaml",
			Run: func(c *Context, _ []string, _ string) error {
				if err := sdkclient.DiagnosticsError(c.Client); err != nil {
					return err
				}
				data, err := c.Client.Diagnostics().OpenAPIYAML(c.Ctx)
				if err != nil {
					return err
				}
				_, err = c.Printer.Stdout().Write(data)
				return err
			},
		},
		{
			Names:    []string{"openapi-yml"},
			Synopsis: "Fetch /openapi.yml.",
			Usage:    "diagnostics openapi-yml",
			Run: func(c *Context, _ []string, _ string) error {
				if err := sdkclient.DiagnosticsError(c.Client); err != nil {
					return err
				}
				data, err := c.Client.Diagnostics().OpenAPIYML(c.Ctx)
				if err != nil {
					return err
				}
				_, err = c.Printer.Stdout().Write(data)
				return err
			},
		},
	}
	r.AddGroup(g)
}
