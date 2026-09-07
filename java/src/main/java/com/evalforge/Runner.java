// EvalForge runner: load an eval set, score cases, aggregate, gate on threshold.
//
// This is the 'unit tests for prompts' engine. In CI, a failing threshold returns a
// non-zero exit code, blocking the deploy -- eval-driven development as a real gate.

package com.evalforge;

import java.io.IOException;
import java.math.BigDecimal;
import java.math.RoundingMode;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.function.Function;
import org.yaml.snakeyaml.Yaml;

public final class Runner {

    private Runner() {}

    public record CaseResult(String id, String scorer, double weight, double score, boolean passed) {}

    public record EvalReport(double aggregate, double threshold, boolean passed, List<CaseResult> cases) {}

    public static EvalReport runEval(Path evalSetPath, Function<String, String> target) throws IOException {
        return runEval(evalSetPath, target, null);
    }

    @SuppressWarnings("unchecked")
    public static EvalReport runEval(Path evalSetPath, Function<String, String> target, Double threshold)
            throws IOException {
        String text = Files.readString(evalSetPath, StandardCharsets.UTF_8);
        Map<String, Object> spec = new Yaml().load(text);

        double effectiveThreshold = threshold != null
            ? threshold
            : (spec.containsKey("threshold") ? toDouble(spec.get("threshold")) : 0.8);

        List<CaseResult> results = new ArrayList<>();
        List<Map<String, Object>> cases = (List<Map<String, Object>>) spec.get("cases");
        for (Map<String, Object> c : cases) {
            String output = target.apply(str(c.get("input")));
            String scorer = str(c.get("scorer"));
            String arg = scorer.equals("regex") ? str(c.get("pattern")) : str(c.get("expected"));
            double score = Scorers.score(scorer, output, arg);
            double weight = c.containsKey("weight") ? toDouble(c.get("weight")) : 1.0;
            double caseThreshold = c.containsKey("pass_score") ? toDouble(c.get("pass_score")) : 0.5;
            results.add(new CaseResult(str(c.get("id")), scorer, weight, score, score >= caseThreshold));
        }

        double totalW = results.stream().mapToDouble(CaseResult::weight).sum();
        if (totalW == 0.0) totalW = 1.0;
        double weighted = results.stream().mapToDouble(r -> r.score() * r.weight()).sum();
        double aggregate = BigDecimal.valueOf(weighted / totalW).setScale(4, RoundingMode.HALF_EVEN).doubleValue();
        return new EvalReport(aggregate, effectiveThreshold, aggregate >= effectiveThreshold, results);
    }

    private static double toDouble(Object v) {
        if (v instanceof Number n) return n.doubleValue();
        return Double.parseDouble(v.toString());
    }

    private static String str(Object v) {
        return v == null ? "" : v.toString();
    }
}
