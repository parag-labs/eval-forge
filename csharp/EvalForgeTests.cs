using System;
using System.IO;
using System.Linq;
using Xunit;

namespace EvalForge.Tests;

public class EvalForgeTests
{
    private static string EvalSet()
    {
        // Walk up from the test output dir to the repo root's examples/eval_set.yaml.
        var dir = AppContext.BaseDirectory;
        while (dir is not null)
        {
            var candidate = Path.Combine(dir, "examples", "eval_set.yaml");
            if (File.Exists(candidate)) return candidate;
            dir = Directory.GetParent(dir)?.FullName;
        }
        throw new FileNotFoundException("examples/eval_set.yaml not found");
    }

    [Fact]
    public void ScorersBasic()
    {
        Assert.Equal(1.0, Scorers.ExactMatch("abc", "abc"));
        Assert.Equal(1.0, Scorers.Contains("the answer is 30 days", "30 days"));
        Assert.Equal(1.0, Scorers.RegexMatch("No discount", "(?i)no"));
        var sem = Scorers.SemanticSimilarity("cancel anytime", "you can cancel anytime");
        Assert.InRange(sem, 0.0, 1.0);
    }

    [Fact]
    public void GoodTargetPassesGate()
    {
        var report = Runner.RunEval(EvalSet(), DemoTarget.Target);
        Assert.True(report.Passed);
        Assert.True(report.Aggregate >= report.Threshold);
    }

    [Fact]
    public void RegressedTargetFailsGate()
    {
        // A regressed agent (drops the refund detail) must fail the eval gate.
        static string Regressed(string question)
        {
            if (question.ToLowerInvariant().Contains("cancel")) return "Yes, cancel anytime.";
            return "I don't know."; // lost refund + discount handling
        }

        var report = Runner.RunEval(EvalSet(), Regressed);
        Assert.False(report.Passed);
    }

    [Fact]
    public void WeightsAffectAggregate()
    {
        var report = Runner.RunEval(EvalSet(), DemoTarget.Target);
        Assert.Contains(report.Cases, c => c.Weight == 2.0);
        Assert.InRange(report.Aggregate, 0.0, 1.0);
    }
}
