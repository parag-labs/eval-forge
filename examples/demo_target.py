"""A tiny demo 'agent' to evaluate. Replace with your real LLM app callable."""


def target(question: str) -> str:
    q = question.lower()
    if "refund" in q:
        return "You have 30 days from purchase to request a full refund."
    if "cancel" in q:
        return "Yes, you can cancel your subscription at any time; it ends at the cycle close."
    if "discount" in q:
        return "No, we do not offer a 90% loyalty discount."
    return "I don't have information on that."
