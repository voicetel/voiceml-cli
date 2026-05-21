package commands

import (
	"strings"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

func registerIncomingPhoneNumbers(r *Registry) {
	g := &Group{Name: "incoming-phone-numbers", Description: "DIDs assigned to the tenant."}
	g.Commands = []*Command{
		{
			Names:    []string{"list"},
			Synopsis: "List incoming phone numbers with optional filters.",
			Usage:    `incoming-phone-numbers list [json]  e.g. {"PhoneNumber":"+18005551234","PageSize":20}`,
			Run: func(c *Context, _ []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				var params voiceml.ListIncomingPhoneNumbersParams
				if err := parseOptionalJSON("incoming-phone-numbers list", tail, &params); err != nil {
					return err
				}
				var p *voiceml.ListIncomingPhoneNumbersParams
				if strings.TrimSpace(tail) != "" {
					p = &params
				}
				out, err := c.Client.IncomingPhoneNumbers().List(c.Ctx, p)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"get"},
			Synopsis: "Fetch one number by SID.",
			Usage:    "incoming-phone-numbers get <sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("incoming-phone-numbers get", args, 1, "<sid>"); err != nil {
					return err
				}
				out, err := c.Client.IncomingPhoneNumbers().Get(c.Ctx, args[0])
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"create"},
			Synopsis: "Assign a DID to the tenant.",
			Usage:    `incoming-phone-numbers create <json>  e.g. {"PhoneNumber":"+18005551234"}`,
			Run: func(c *Context, _ []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				var body voiceml.CreateIncomingPhoneNumberParams
				if err := parseJSON("incoming-phone-numbers create", tail, &body); err != nil {
					return err
				}
				out, err := c.Client.IncomingPhoneNumbers().Create(c.Ctx, body)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"update"},
			Synopsis: "Update voice routing on a DID.",
			Usage:    `incoming-phone-numbers update <sid> <json>`,
			Run: func(c *Context, args []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("incoming-phone-numbers update", args, 1, "<sid> <json>"); err != nil {
					return err
				}
				var body voiceml.UpdateIncomingPhoneNumberParams
				if err := parseJSON("incoming-phone-numbers update", skipArgs(tail, 1), &body); err != nil {
					return err
				}
				out, err := c.Client.IncomingPhoneNumbers().Update(c.Ctx, args[0], body)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"delete"},
			Synopsis: "Release a DID from the tenant.",
			Usage:    "incoming-phone-numbers delete <sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("incoming-phone-numbers delete", args, 1, "<sid>"); err != nil {
					return err
				}
				if err := c.Client.IncomingPhoneNumbers().Delete(c.Ctx, args[0]); err != nil {
					return err
				}
				c.Printer.Println("Deleted.")
				return nil
			},
		},
	}
	r.AddGroup(g)
}
