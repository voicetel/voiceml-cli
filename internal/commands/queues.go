package commands

import voiceml "github.com/voicetel/voiceml-go-sdk"

func registerQueues(r *Registry) {
	g := &Group{Name: "queues", Description: "Call queues, members, dequeue."}
	g.Commands = []*Command{
		{
			Names:    []string{"list"},
			Synopsis: "List queues.",
			Usage:    "queues list",
			Run: func(c *Context, _ []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				out, err := c.Client.Queues().List(c.Ctx)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"get"},
			Synopsis: "Fetch one queue by SID.",
			Usage:    "queues get <queue_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("queues get", args, 1, "<queue_sid>"); err != nil {
					return err
				}
				out, err := c.Client.Queues().Get(c.Ctx, args[0])
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"create"},
			Synopsis: "Create a queue.",
			Usage:    `queues create <json>  e.g. {"FriendlyName":"support","MaxSize":100}`,
			Run: func(c *Context, _ []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				var body voiceml.CreateQueueParams
				if err := parseJSON("queues create", tail, &body); err != nil {
					return err
				}
				out, err := c.Client.Queues().Create(c.Ctx, body)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"update"},
			Synopsis: "Update a queue.",
			Usage:    `queues update <queue_sid> <json>`,
			Run: func(c *Context, args []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("queues update", args, 1, "<queue_sid> <json>"); err != nil {
					return err
				}
				var body voiceml.UpdateQueueParams
				if err := parseJSON("queues update", skipArgs(tail, 1), &body); err != nil {
					return err
				}
				out, err := c.Client.Queues().Update(c.Ctx, args[0], body)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"delete"},
			Synopsis: "Delete a queue.",
			Usage:    "queues delete <queue_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("queues delete", args, 1, "<queue_sid>"); err != nil {
					return err
				}
				if err := c.Client.Queues().Delete(c.Ctx, args[0]); err != nil {
					return err
				}
				c.Printer.Println("Deleted.")
				return nil
			},
		},
		{
			Names:    []string{"list-members"},
			Synopsis: "List members waiting in a queue.",
			Usage:    `queues list-members <queue_sid> [json]`,
			Run: func(c *Context, args []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("queues list-members", args, 1, "<queue_sid> [json]"); err != nil {
					return err
				}
				var params voiceml.ListQueueMembersParams
				if err := parseOptionalJSON("queues list-members", skipArgs(tail, 1), &params); err != nil {
					return err
				}
				out, err := c.Client.Queues().ListMembers(c.Ctx, args[0], params)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"peek-front"},
			Synopsis: "Peek at the front queue member.",
			Usage:    "queues peek-front <queue_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("queues peek-front", args, 1, "<queue_sid>"); err != nil {
					return err
				}
				out, err := c.Client.Queues().PeekFront(c.Ctx, args[0])
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"dequeue-front"},
			Synopsis: "Dequeue the front member.",
			Usage:    `queues dequeue-front <queue_sid> <json>  e.g. {"Url":"https://example.com/twiml"}`,
			Run: func(c *Context, args []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("queues dequeue-front", args, 1, "<queue_sid> <json>"); err != nil {
					return err
				}
				var body voiceml.DequeueParams
				if err := parseJSON("queues dequeue-front", skipArgs(tail, 1), &body); err != nil {
					return err
				}
				out, err := c.Client.Queues().DequeueFront(c.Ctx, args[0], body)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"get-member"},
			Synopsis: "Fetch a specific queue member.",
			Usage:    "queues get-member <queue_sid> <call_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("queues get-member", args, 2, "<queue_sid> <call_sid>"); err != nil {
					return err
				}
				out, err := c.Client.Queues().GetMember(c.Ctx, args[0], args[1])
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"dequeue-member"},
			Synopsis: "Dequeue a specific member.",
			Usage:    `queues dequeue-member <queue_sid> <call_sid> <json>`,
			Run: func(c *Context, args []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("queues dequeue-member", args, 2, "<queue_sid> <call_sid> <json>"); err != nil {
					return err
				}
				var body voiceml.DequeueParams
				if err := parseJSON("queues dequeue-member", skipArgs(tail, 2), &body); err != nil {
					return err
				}
				out, err := c.Client.Queues().DequeueMember(c.Ctx, args[0], args[1], body)
				return printResultG(c, out, err)
			},
		},
	}
	r.AddGroup(g)
}
