package com.evalforge;

import static org.junit.jupiter.api.Assertions.*;

import com.evalforge.Runner.EvalReport;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import org.junit.jupiter.api.Test;

class EvalForgeTest {

    private static Path evalSet() {
        // Walk up from the module dir to the repo root's examples/eval_set.yaml.
        Path dir = Paths.get("").toAbsolutePath();
        while (dir != null) {
            Path candidate = dir.resolve("examples").resolve("eval_set.yaml");
            if (Files.exists(candidate)) return candidate;
            dir = dir.getParent();
        }
        throw new IllegalStateException("examples/eval_set.yaml not found");
    }

    @Test
    void scorersBasic() {
        assertEquals(1.0, Scorers.exactMatch("abc", "abc"));
        assertEquals(1.0, Scorers.contains("the answer is 30 days", "30 days"));
        assertEquals(1.0, Scorers.regexMatch("No discount", "(?i)no"));
        double sem = Scorers.semanticSimilarity("cancel anytime", "you can cancel anytime");
        assertTrue(sem >= 0.0 && sem <= 1.0);
    }

    @Test
    void goodTargetPassesGate() throws IOException {
        EvalReport report = Runner.runEval(evalSet(), DemoTarget::target);
        assertTrue(report.passed());
        assertTrue(report.aggregate() >= report.threshold());
    }

    @Test
    void regressedTargetFailsGate() throws IOException {
        // A regressed agent (drops the refund detail) must fail the eval gate.
        EvalReport report = Runner.runEval(evalSet(), question -> {
            if (question.toLowerCase().contains("cancel")) return "Yes, cancel anytime.";
            return "I don't know."; // lost refund + discount handling
        });
        assertFalse(report.passed());
    }

    @Test
    void weightsAffectAggregate() throws IOException {
        EvalReport report = Runner.runEval(evalSet(), DemoTarget::target);
        assertTrue(report.cases().stream().anyMatch(c -> c.weight() == 2.0));
        assertTrue(report.aggregate() >= 0.0 && report.aggregate() <= 1.0);
    }
}
