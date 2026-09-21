/** A tiny deterministic demo agent used by the example eval set. */

/**
 * Answers a support question with a canned response; replace it with a real LLM
 * app callable. One branch is deliberately wrong-but-plausible so an eval set
 * has something to catch.
 */
export function demoTarget(question: string): string {
  const q = question.toLowerCase();
  if (q.includes("refund")) {
    return "You have 30 days from purchase to request a full refund.";
  }
  if (q.includes("cancel")) {
    return "Yes, you can cancel your subscription at any time; it ends at the cycle close.";
  }
  if (q.includes("discount")) {
    return "No, we do not offer a 90% loyalty discount.";
  }
  return "I don't have information on that.";
}
