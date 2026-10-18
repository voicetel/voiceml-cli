package repl

// RecordHistory permits only recognized, read-only commands. Credential entry,
// payment data, mutations, unknown commands and malformed input are never saved.
func RecordHistory(line string) bool {
	p, err := Parse(line)
	if err != nil || len(p.Tokens) == 0 {
		return false
	}
	switch p.Tokens[0] {
	case "help", "whoami", "clear", "exit", "quit", "completion":
		return len(p.Tokens) <= 2
	case "calls", "conferences", "queues", "applications", "recordings", "incoming-phone-numbers", "messages", "diagnostics":
		if len(p.Tokens) < 2 {
			return false
		}
		switch p.Tokens[1] {
		case "list", "get", "list-participants", "get-participant", "list-recordings", "list-members", "get-member", "peek-front", "health", "openapi", "openapi-yaml", "openapi-yml":
			return true
		}
	}
	return false
}
