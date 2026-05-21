package commands

import voiceml "github.com/voicetel/voiceml-go-sdk"

func registerCalls(r *Registry) {
	g := &Group{Name: "calls", Description: "Outbound/inbound calls, AMD, live updates."}
	g.Commands = []*Command{
		{
			Names:    []string{"list"},
			Synopsis: "List calls with optional filters.",
			Usage:    `calls list [json]  e.g. {"Status":"completed","PageSize":20}`,
			Run: func(c *Context, _ []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				var params voiceml.ListCallsParams
				if err := parseOptionalJSON("calls list", tail, &params); err != nil {
					return err
				}
				out, err := c.Client.Calls().List(c.Ctx, params)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"get"},
			Synopsis: "Fetch one call by SID.",
			Usage:    "calls get <call_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("calls get", args, 1, "<call_sid>"); err != nil {
					return err
				}
				out, err := c.Client.Calls().Get(c.Ctx, args[0])
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"create"},
			Synopsis: "Originate a new outbound call.",
			Usage:    `calls create <json>  e.g. {"To":"+18005551234","From":"+18005550000","Url":"https://example.com/twiml"}`,
			Run: func(c *Context, _ []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				var body voiceml.CreateCallParams
				if err := parseJSON("calls create", tail, &body); err != nil {
					return err
				}
				out, err := c.Client.Calls().Create(c.Ctx, body)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"update"},
			Synopsis: "Update or terminate a live call.",
			Usage:    `calls update <call_sid> <json>  e.g. {"Status":"completed"}`,
			Run: func(c *Context, args []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("calls update", args, 1, "<call_sid> <json>"); err != nil {
					return err
				}
				var body voiceml.UpdateCallParams
				if err := parseJSON("calls update", skipArgs(tail, 1), &body); err != nil {
					return err
				}
				out, err := c.Client.Calls().Update(c.Ctx, args[0], body)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"delete"},
			Synopsis: "Delete a completed call record.",
			Usage:    "calls delete <call_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("calls delete", args, 1, "<call_sid>"); err != nil {
					return err
				}
				if err := c.Client.Calls().Delete(c.Ctx, args[0]); err != nil {
					return err
				}
				c.Printer.Println("Deleted.")
				return nil
			},
		},
	}
	r.AddGroup(g)
}
