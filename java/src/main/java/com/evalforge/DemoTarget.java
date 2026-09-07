// A tiny demo 'agent' to evaluate. Replace with your real LLM app callable.

package com.evalforge;

public final class DemoTarget {

    private DemoTarget() {}

    public static String target(String question) {
        String q = question.toLowerCase();
        if (q.contains("refund")) {
            return "You have 30 days from purchase to request a full refund.";
        }
        if (q.contains("cancel")) {
            return "Yes, you can cancel your subscription at any time; it ends at the cycle close.";
        }
        if (q.contains("discount")) {
            return "No, we do not offer a 90% loyalty discount.";
        }
        return "I don't have information on that.";
    }
}
