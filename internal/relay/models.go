package relay

// MainstreamProbeModels are the flagship models we evaluate across relays.
// Domestic models (e.g. DeepSeek) and legacy IDs (gpt-4o, sonnet-4.5, gemini-2.0) are excluded.
var MainstreamProbeModels = []string{
	"gpt-5.5",
	"gpt-5.5-pro",
	"claude-opus-4-7",
	"gemini-3.1-pro-preview", // Gemini 3.x flagship; API id on most relays
}

// DefaultHealthModel is used for lightweight availability checks.
const DefaultHealthModel = "gpt-5.4-mini"
