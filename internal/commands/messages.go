package commands

import voiceml "github.com/voicetel/voiceml-go-sdk"

func registerMessages(r *Registry) {
	g := &Group{Name: "messages", Description: "Twilio-compatible outbound SMS dispatch and metadata."}
	g.Commands = []*Command{
		{
			Names:    []string{"list"},
			Synopsis: "List messages with optional filters.",
			Usage:    `messages list [json]  e.g. {"To":"+18005551234","PageSize":20}`,
			Run: func(c *Context, _ []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				var params voiceml.ListMessagesParams
				if err := parseOptionalJSON("messages list", tail, &params); err != nil {
					return err
				}
				out, err := c.Client.Messages().List(c.Ctx, params)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"get"},
			Synopsis: "Fetch one message by SID.",
			Usage:    "messages get <message_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("messages get", args, 1, "<message_sid>"); err != nil {
					return err
				}
				out, err := c.Client.Messages().Fetch(c.Ctx, args[0])
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"create"},
			Synopsis: "Dispatch a new outbound SMS.",
			Usage:    `messages create <json>  e.g. {"To":"+18005551234","Body":"hello","From":"+18005550000"}`,
			Run: func(c *Context, _ []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				var body voiceml.CreateMessageParams
				if err := parseJSON("messages create", tail, &body); err != nil {
					return err
				}
				out, err := c.Client.Messages().Create(c.Ctx, body)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"update"},
			Synopsis: "Update a message (redact Body or attempt cancel).",
			Usage:    `messages update <message_sid> [json]  e.g. {"Body":""}`,
			Run: func(c *Context, args []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("messages update", args, 1, "<message_sid> [json]"); err != nil {
					return err
				}
				var body voiceml.UpdateMessageParams
				if err := parseOptionalJSON("messages update", skipArgs(tail, 1), &body); err != nil {
					return err
				}
				out, err := c.Client.Messages().Update(c.Ctx, args[0], body)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"delete"},
			Synopsis: "Delete a message resource.",
			Usage:    "messages delete <message_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("messages delete", args, 1, "<message_sid>"); err != nil {
					return err
				}
				if err := c.Client.Messages().Delete(c.Ctx, args[0]); err != nil {
					return err
				}
				c.Printer.Println("Deleted.")
				return nil
			},
		},
	}
	r.AddGroup(g)
}
