// EvalForge runner: load an eval set, score cases, aggregate, gate on threshold.
//
// This is the 'unit tests for prompts' engine. In CI, a failing threshold returns a
// non-zero exit code, blocking the deploy -- eval-driven development as a real gate.

using System;
using System.Collections.Generic;
using System.Globalization;
using System.IO;
using System.Linq;
using YamlDotNet.Serialization;

namespace EvalForge;

public sealed record CaseResult(string Id, string Scorer, double Weight, double Score, bool Passed);

public sealed record EvalReport(double Aggregate, double Threshold, bool Passed, List<CaseResult> Cases);

public static class Runner
{
    public static EvalReport RunEval(string evalSetPath, Func<string, string> target, double? threshold = null)
    {
        var text = File.ReadAllText(evalSetPath);
        var deserializer = new DeserializerBuilder().Build();
        var spec = deserializer.Deserialize<Dictionary<string, object>>(text);

        var effectiveThreshold = threshold
            ?? (spec.TryGetValue("threshold", out var t) ? ToDouble(t) : 0.8);

        var results = new List<CaseResult>();
        var cases = (List<object>)spec["cases"];
        foreach (var caseObj in cases)
        {
            var c = ToStringMap(caseObj);
            var output = target(c["input"]);
            var scorer = c["scorer"];
            var arg = scorer == "regex" ? c["pattern"] : c["expected"];
            var score = Scorers.Score(scorer, output, arg);
            var weight = c.TryGetValue("weight", out var w) ? ToDouble(w) : 1.0;
            var caseThreshold = c.TryGetValue("pass_score", out var ps) ? ToDouble(ps) : 0.5;
            results.Add(new CaseResult(c["id"], scorer, weight, score, score >= caseThreshold));
        }

        var totalW = results.Sum(r => r.Weight);
        if (totalW == 0.0) totalW = 1.0;
        var aggregate = Math.Round(results.Sum(r => r.Score * r.Weight) / totalW, 4);
        return new EvalReport(aggregate, effectiveThreshold, aggregate >= effectiveThreshold, results);
    }

    private static double ToDouble(object v)
        => double.Parse(v.ToString()!, CultureInfo.InvariantCulture);

    private static Dictionary<string, string> ToStringMap(object caseObj)
    {
        var map = new Dictionary<string, string>();
        foreach (var kv in (Dictionary<object, object>)caseObj)
            map[kv.Key.ToString()!] = kv.Value?.ToString() ?? "";
        return map;
    }
}
