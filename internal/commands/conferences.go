package commands

import voiceml "github.com/voicetel/voiceml-go-sdk"

func registerConferences(r *Registry) {
	g := &Group{Name: "conferences", Description: "Conference rooms, participants, recordings."}
	g.Commands = []*Command{
		{
			Names:    []string{"list"},
			Synopsis: "List conferences with optional filters.",
			Usage:    `conferences list [json]  e.g. {"Status":"in-progress","PageSize":20}`,
			Run: func(c *Context, _ []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				var params voiceml.ListConferencesParams
				if err := parseOptionalJSON("conferences list", tail, &params); err != nil {
					return err
				}
				out, err := c.Client.Conferences().List(c.Ctx, params)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"get"},
			Synopsis: "Fetch one conference by SID.",
			Usage:    "conferences get <conference_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("conferences get", args, 1, "<conference_sid>"); err != nil {
					return err
				}
				out, err := c.Client.Conferences().Get(c.Ctx, args[0])
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"end"},
			Synopsis: "End a conference.",
			Usage:    `conferences end <conference_sid> [json]  default {"Status":"completed"}`,
			Run: func(c *Context, args []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("conferences end", args, 1, "<conference_sid>"); err != nil {
					return err
				}
				var params *voiceml.EndConferenceParams
				if stringsTrim := skipArgs(tail, 1); stringsTrim != "" {
					var body voiceml.EndConferenceParams
					if err := parseJSON("conferences end", stringsTrim, &body); err != nil {
						return err
					}
					params = &body
				}
				out, err := c.Client.Conferences().End(c.Ctx, args[0], params)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"list-participants"},
			Synopsis: "List participants in a conference.",
			Usage:    `conferences list-participants <conference_sid> [json]`,
			Run: func(c *Context, args []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("conferences list-participants", args, 1, "<conference_sid> [json]"); err != nil {
					return err
				}
				var params voiceml.ListParticipantsParams
				if err := parseOptionalJSON("conferences list-participants", skipArgs(tail, 1), &params); err != nil {
					return err
				}
				out, err := c.Client.Conferences().ListParticipants(c.Ctx, args[0], params)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"get-participant"},
			Synopsis: "Fetch one participant.",
			Usage:    "conferences get-participant <conference_sid> <call_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("conferences get-participant", args, 2, "<conference_sid> <call_sid>"); err != nil {
					return err
				}
				out, err := c.Client.Conferences().GetParticipant(c.Ctx, args[0], args[1])
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"update-participant"},
			Synopsis: "Mute/unmute or hold/unhold a participant.",
			Usage:    `conferences update-participant <conference_sid> <call_sid> <json>`,
			Run: func(c *Context, args []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("conferences update-participant", args, 2, "<conference_sid> <call_sid> <json>"); err != nil {
					return err
				}
				var body voiceml.UpdateParticipantParams
				if err := parseJSON("conferences update-participant", skipArgs(tail, 2), &body); err != nil {
					return err
				}
				out, err := c.Client.Conferences().UpdateParticipant(c.Ctx, args[0], args[1], body)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"kick-participant"},
			Synopsis: "Remove a participant from a conference.",
			Usage:    "conferences kick-participant <conference_sid> <call_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("conferences kick-participant", args, 2, "<conference_sid> <call_sid>"); err != nil {
					return err
				}
				if err := c.Client.Conferences().KickParticipant(c.Ctx, args[0], args[1]); err != nil {
					return err
				}
				c.Printer.Println("Participant kicked.")
				return nil
			},
		},
		{
			Names:    []string{"list-recordings"},
			Synopsis: "List recordings for a conference.",
			Usage:    "conferences list-recordings <conference_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("conferences list-recordings", args, 1, "<conference_sid>"); err != nil {
					return err
				}
				out, err := c.Client.Conferences().ListRecordings(c.Ctx, args[0], voiceml.ListCallRecordingsParams{})
				return printResultG(c, out, err)
			},
		},
	}
	r.AddGroup(g)
}
