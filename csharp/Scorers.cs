// EvalForge scorers: turn (output, expected) into a 0..1 score.

using System;
using System.Collections.Generic;
using System.Linq;
using System.Text.RegularExpressions;

namespace EvalForge;

public static class Scorers
{
    public static double ExactMatch(string output, string expected)
        => output.Trim() == expected.Trim() ? 1.0 : 0.0;

    public static double Contains(string output, string expected)
        => output.ToLowerInvariant().Contains(expected.Trim().ToLowerInvariant()) ? 1.0 : 0.0;

    public static double RegexMatch(string output, string pattern)
        => Regex.IsMatch(output, pattern) ? 1.0 : 0.0;

    /// <summary>
    /// Lightweight, dependency-free proxy for semantic closeness: token overlap +
    /// sequence ratio (Ratcliff/Obershelp, matching Python's difflib). Swap for
    /// embedding cosine or an LLM-as-judge in production; the interface is identical.
    /// </summary>
    public static double SemanticSimilarity(string output, string expected)
    {
        var a = output.ToLowerInvariant().Split((char[]?)null, StringSplitOptions.RemoveEmptyEntries);
        var b = expected.ToLowerInvariant().Split((char[]?)null, StringSplitOptions.RemoveEmptyEntries);
        if (a.Length == 0 || b.Length == 0) return 0.0;
        var sa = new HashSet<string>(a);
        var sb = new HashSet<string>(b);
        var inter = sa.Count(x => sb.Contains(x));
        var union = new HashSet<string>(sa);
        union.UnionWith(sb);
        var overlap = (double)inter / union.Count;
        var ratio = SequenceRatio(output.ToLowerInvariant(), expected.ToLowerInvariant());
        return Math.Round(0.5 * overlap + 0.5 * ratio, 4);
    }

    // --- difflib.SequenceMatcher.ratio() (Ratcliff/Obershelp), no autojunk ---

    private static double SequenceRatio(string a, string b)
    {
        if (a.Length + b.Length == 0) return 1.0;
        var b2j = new Dictionary<char, List<int>>();
        for (var j = 0; j < b.Length; j++)
        {
            if (!b2j.TryGetValue(b[j], out var list)) { list = new List<int>(); b2j[b[j]] = list; }
            list.Add(j);
        }
        var matches = TotalMatches(a, b, b2j, 0, a.Length, 0, b.Length);
        return 2.0 * matches / (a.Length + b.Length);
    }

    private static int TotalMatches(string a, string b, Dictionary<char, List<int>> b2j,
        int alo, int ahi, int blo, int bhi)
    {
        var (i, j, k) = FindLongestMatch(a, b2j, alo, ahi, blo, bhi);
        if (k == 0) return 0;
        return k
            + TotalMatches(a, b, b2j, alo, i, blo, j)
            + TotalMatches(a, b, b2j, i + k, ahi, j + k, bhi);
    }

    private static (int, int, int) FindLongestMatch(string a, Dictionary<char, List<int>> b2j,
        int alo, int ahi, int blo, int bhi)
    {
        int besti = alo, bestj = blo, bestsize = 0;
        var j2len = new Dictionary<int, int>();
        for (var i = alo; i < ahi; i++)
        {
            var newj2len = new Dictionary<int, int>();
            if (b2j.TryGetValue(a[i], out var js))
            {
                foreach (var j in js)
                {
                    if (j < blo) continue;
                    if (j >= bhi) break;
                    var k = (j2len.TryGetValue(j - 1, out var prev) ? prev : 0) + 1;
                    newj2len[j] = k;
                    if (k > bestsize)
                    {
                        besti = i - k + 1;
                        bestj = j - k + 1;
                        bestsize = k;
                    }
                }
            }
            j2len = newj2len;
        }
        return (besti, bestj, bestsize);
    }

    private static readonly HashSet<string> Names = new() { "exact_match", "contains", "regex", "semantic" };

    public static bool IsKnown(string name) => Names.Contains(name);

    /// <summary>Score by scorer name. <paramref name="arg"/> is the pattern for
    /// "regex", otherwise the expected string.</summary>
    public static double Score(string name, string output, string arg) => name switch
    {
        "exact_match" => ExactMatch(output, arg),
        "contains" => Contains(output, arg),
        "regex" => RegexMatch(output, arg),
        "semantic" => SemanticSimilarity(output, arg),
        _ => throw new ArgumentException(
            $"unknown scorer '{name}'. available: [{string.Join(", ", Names)}]"),
    };
}
