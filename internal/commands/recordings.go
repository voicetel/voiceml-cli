package commands

import (
	"encoding/base64"
	"os"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

func registerRecordings(r *Registry) {
	g := &Group{Name: "recordings", Description: "Account-wide recordings, metadata, audio."}
	g.Commands = []*Command{
		{
			Names:    []string{"list"},
			Synopsis: "List recordings with optional pagination.",
			Usage:    `recordings list [json]  e.g. {"PageSize":20}`,
			Run: func(c *Context, _ []string, tail string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				var params voiceml.ListRecordingsParams
				if err := parseOptionalJSON("recordings list", tail, &params); err != nil {
					return err
				}
				out, err := c.Client.Recordings().List(c.Ctx, params)
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"get"},
			Synopsis: "Fetch recording metadata.",
			Usage:    "recordings get <recording_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("recordings get", args, 1, "<recording_sid>"); err != nil {
					return err
				}
				out, err := c.Client.Recordings().Get(c.Ctx, args[0])
				return printResultG(c, out, err)
			},
		},
		{
			Names:    []string{"get-audio"},
			Synopsis: "Fetch recording audio (WAV).",
			Usage: "recordings get-audio <recording_sid> [output-file]\n" +
				"  Without output-file, prints base64-encoded JSON. With a path, writes raw WAV bytes.",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("recordings get-audio", args, 1, "<recording_sid> [output-file]"); err != nil {
					return err
				}
				data, contentType, err := c.Client.Recordings().GetAudio(c.Ctx, args[0])
				if err != nil {
					return err
				}
				if len(args) >= 2 {
					if err := os.WriteFile(args[1], data, 0o600); err != nil {
						return err
					}
					c.Printer.Printf("Wrote %d bytes to %s\n", len(data), args[1])
					return nil
				}
				return c.Printer.JSON(map[string]any{
					"contentType": contentType,
					"size":        len(data),
					"audioBase64": base64.StdEncoding.EncodeToString(data),
				})
			},
		},
		{
			Names:    []string{"delete"},
			Synopsis: "Delete a recording.",
			Usage:    "recordings delete <recording_sid>",
			Run: func(c *Context, args []string, _ string) error {
				if err := requireConfigured(c); err != nil {
					return err
				}
				if err := requireArgs("recordings delete", args, 1, "<recording_sid>"); err != nil {
					return err
				}
				if err := c.Client.Recordings().Delete(c.Ctx, args[0]); err != nil {
					return err
				}
				c.Printer.Println("Deleted.")
				return nil
			},
		},
	}
	r.AddGroup(g)
}
