package commands

import voiceml "github.com/voicetel/voiceml-go-sdk"

func registerApplications(r *Registry) {
	g := &Group{Name: "applications", Description: "Stored TwiML + callback bundles."}
	g.Commands = []*Command{
		{
			Names:    []string{"list"},
			Synopsis: "List applications with optional filters.",
			Usage:    `applications list [json]  e.g. {"FriendlyName":"ivr","PageSize":20}`,
			Run: func(c *Context, _ []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				var params voiceml.ListApplicationsParams
				if err := parseOptionalJSON("applications list", tail, &params); err != nil {
					return err
				}
				out, err := c.Client.Applications().List(c.Ctx, params)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"get"},
			Synopsis: "Fetch one application by SID.",
			Usage:    "applications get <application_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("applications get", args, 1, "<application_sid>"); err != nil {
					return err
				}
				out, err := c.Client.Applications().Get(c.Ctx, args[0])
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"create"},
			Synopsis: "Create an application.",
			Usage:    `applications create <json>  e.g. {"FriendlyName":"ivr","VoiceUrl":"https://example.com/voice"}`,
			Run: func(c *Context, _ []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				var body voiceml.ApplicationParams
				if err := parseJSON("applications create", tail, &body); err != nil {
					return err
				}
				out, err := c.Client.Applications().Create(c.Ctx, body)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"update"},
			Synopsis: "Update an application.",
			Usage:    `applications update <application_sid> <json>`,
			Run: func(c *Context, args []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("applications update", args, 1, "<application_sid> <json>"); err != nil {
					return err
				}
				var body voiceml.ApplicationParams
				if err := parseJSON("applications update", skipArgs(tail, 1), &body); err != nil {
					return err
				}
				out, err := c.Client.Applications().Update(c.Ctx, args[0], body)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"delete"},
			Synopsis: "Delete an application.",
			Usage:    "applications delete <application_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("applications delete", args, 1, "<application_sid>"); err != nil {
					return err
				}
				if err := c.Client.Applications().Delete(c.Ctx, args[0]); err != nil {
					return err
				}
				c.Printer.Println("Deleted.")
				return nil
			},
		},
	}
	r.AddGroup(g)
}
