// EvalForge scorers: turn (output, expected) into a 0..1 score.

package com.evalforge;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.regex.Pattern;

public final class Scorers {

    private Scorers() {}

    public static double exactMatch(String output, String expected) {
        return output.strip().equals(expected.strip()) ? 1.0 : 0.0;
    }

    public static double contains(String output, String expected) {
        return output.toLowerCase().contains(expected.strip().toLowerCase()) ? 1.0 : 0.0;
    }

    public static double regexMatch(String output, String pattern) {
        return Pattern.compile(pattern).matcher(output).find() ? 1.0 : 0.0;
    }

    /** Lightweight, dependency-free proxy for semantic closeness: token overlap +
     * sequence ratio (Ratcliff/Obershelp, matching Python's difflib). Swap for
     * embedding cosine or an LLM-as-judge in production; the interface is identical. */
    public static double semanticSimilarity(String output, String expected) {
        String[] a = output.toLowerCase().split("\\s+");
        String[] b = expected.toLowerCase().split("\\s+");
        List<String> al = nonEmpty(a);
        List<String> bl = nonEmpty(b);
        if (al.isEmpty() || bl.isEmpty()) return 0.0;
        Set<String> sa = new HashSet<>(al);
        Set<String> sb = new HashSet<>(bl);
        Set<String> inter = new HashSet<>(sa);
        inter.retainAll(sb);
        Set<String> union = new HashSet<>(sa);
        union.addAll(sb);
        double overlap = (double) inter.size() / union.size();
        double ratio = sequenceRatio(output.toLowerCase(), expected.toLowerCase());
        return round4(0.5 * overlap + 0.5 * ratio);
    }

    private static List<String> nonEmpty(String[] parts) {
        List<String> out = new ArrayList<>();
        for (String p : parts) if (!p.isEmpty()) out.add(p);
        return out;
    }

    private static double round4(double v) {
        return BigDecimal.valueOf(v).setScale(4, RoundingMode.HALF_EVEN).doubleValue();
    }

    // --- difflib.SequenceMatcher.ratio() (Ratcliff/Obershelp), no autojunk ---

    private static double sequenceRatio(String a, String b) {
        if (a.length() + b.length() == 0) return 1.0;
        Map<Character, List<Integer>> b2j = new HashMap<>();
        for (int j = 0; j < b.length(); j++) {
            b2j.computeIfAbsent(b.charAt(j), k -> new ArrayList<>()).add(j);
        }
        int matches = totalMatches(a, b2j, 0, a.length(), 0, b.length());
        return 2.0 * matches / (a.length() + b.length());
    }

    private static int totalMatches(String a, Map<Character, List<Integer>> b2j,
                                    int alo, int ahi, int blo, int bhi) {
        int[] m = findLongestMatch(a, b2j, alo, ahi, blo, bhi);
        int i = m[0], j = m[1], k = m[2];
        if (k == 0) return 0;
        return k
            + totalMatches(a, b2j, alo, i, blo, j)
            + totalMatches(a, b2j, i + k, ahi, j + k, bhi);
    }

    private static int[] findLongestMatch(String a, Map<Character, List<Integer>> b2j,
                                          int alo, int ahi, int blo, int bhi) {
        int besti = alo, bestj = blo, bestsize = 0;
        Map<Integer, Integer> j2len = new HashMap<>();
        for (int i = alo; i < ahi; i++) {
            Map<Integer, Integer> newj2len = new HashMap<>();
            List<Integer> js = b2j.get(a.charAt(i));
            if (js != null) {
                for (int j : js) {
                    if (j < blo) continue;
                    if (j >= bhi) break;
                    int k = j2len.getOrDefault(j - 1, 0) + 1;
                    newj2len.put(j, k);
                    if (k > bestsize) {
                        besti = i - k + 1;
                        bestj = j - k + 1;
                        bestsize = k;
                    }
                }
            }
            j2len = newj2len;
        }
        return new int[] {besti, bestj, bestsize};
    }

    private static final Set<String> NAMES = Set.of("exact_match", "contains", "regex", "semantic");

    public static boolean isKnown(String name) {
        return NAMES.contains(name);
    }

    /** Score by scorer name. {@code arg} is the pattern for "regex", otherwise the
     * expected string. */
    public static double score(String name, String output, String arg) {
        return switch (name) {
            case "exact_match" -> exactMatch(output, arg);
            case "contains" -> contains(output, arg);
            case "regex" -> regexMatch(output, arg);
            case "semantic" -> semanticSimilarity(output, arg);
            default -> throw new IllegalArgumentException(
                "unknown scorer '" + name + "'. available: " + NAMES);
        };
    }
}
