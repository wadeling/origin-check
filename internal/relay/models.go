package relay

// MainstreamProbeModels are the flagship models we evaluate across relays.
// Legacy IDs (gpt-4o, sonnet-4.x, gemini-2.x, deepseek, etc.) are excluded.
var MainstreamProbeModels = []string{
	"gpt-5.5",
	"claude-opus-4-7",
	"gemini-3.5-flash",
}

// DefaultHealthModel is used for lightweight availability checks.
const DefaultHealthModel = "gpt-5.4-mini"

// LegacyModels are no longer probed; reports for these may be purged on seed.
var LegacyModels = []string{
	"gpt-4o", "gpt-4o-mini", "gpt-4", "gpt-3.5-turbo",
	"claude-sonnet-4", "claude-sonnet-4-5", "claude-sonnet-4-6",
	"gpt-5.5-pro",
	"gemini-2.0-flash", "gemini-2.0-pro", "gemini-2.5-flash",
	"gemini-3.1-pro-preview",
	"deepseek-chat", "deepseek-v3",
}
