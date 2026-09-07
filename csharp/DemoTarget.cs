// A tiny demo 'agent' to evaluate. Replace with your real LLM app callable.

namespace EvalForge;

public static class DemoTarget
{
    public static string Target(string question)
    {
        var q = question.ToLowerInvariant();
        if (q.Contains("refund"))
            return "You have 30 days from purchase to request a full refund.";
        if (q.Contains("cancel"))
            return "Yes, you can cancel your subscription at any time; it ends at the cycle close.";
        if (q.Contains("discount"))
            return "No, we do not offer a 90% loyalty discount.";
        return "I don't have information on that.";
    }
}
