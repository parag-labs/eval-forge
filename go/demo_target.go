package evalforge

import "strings"

// DemoTarget is a tiny demo "agent" to evaluate; replace it with a real LLM app
// callable. It answers a few canned support questions and otherwise declines.
func DemoTarget(question string) string {
	q := strings.ToLower(question)
	switch {
	case strings.Contains(q, "refund"):
		return "You have 30 days from purchase to request a full refund."
	case strings.Contains(q, "cancel"):
		return "Yes, you can cancel your subscription at any time; it ends at the cycle close."
	case strings.Contains(q, "discount"):
		return "No, we do not offer a 90% loyalty discount."
	default:
		return "I don't have information on that."
	}
}
