import { existsSync, rmSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { afterAll, describe, expect, it } from "vitest";
import { demoTarget } from "./demo_target";
import { runEval } from "./runner";

function evalSetPath(): string {
  let dir = process.cwd();
  for (;;) {
    const candidate = join(dir, "examples", "eval_set.yaml");
    if (existsSync(candidate)) return candidate;
    const parent = dirname(dir);
    if (parent === dir) throw new Error("examples/eval_set.yaml not found walking up from cwd");
    dir = parent;
  }
}

let tempCounter = 0;
const tempFiles: string[] = [];

/** Writes an eval set into the package directory (never a system temp dir) and
 * schedules its removal after the suite. */
function writeTempEval(content: string): string {
  const name = join(process.cwd(), `eval-test-${process.pid}-${tempCounter++}.yaml`);
  writeFileSync(name, content, "utf8");
  tempFiles.push(name);
  return name;
}

afterAll(() => {
  for (const f of tempFiles) {
    try {
      rmSync(f);
    } catch {
      // best-effort cleanup
    }
  }
});

describe("runEval", () => {
  it("passes the gate for the good demo target", () => {
    const report = runEval(evalSetPath(), demoTarget);
    expect(report.passed).toBe(true);
    expect(report.aggregate).toBeGreaterThanOrEqual(report.threshold);
  });

  it("produces the exact reference aggregate and threshold", () => {
    const report = runEval(evalSetPath(), demoTarget);
    expect(report.aggregate).toBeCloseTo(0.9238, 10);
    expect(report.threshold).toBe(0.75);
  });

  it("reports three cases in eval-set order", () => {
    const report = runEval(evalSetPath(), demoTarget);
    expect(report.cases.map((c) => c.id)).toEqual([
      "refund-window",
      "cancellation",
      "no-hallucinated-discount",
    ]);
  });

  it("carries case weights and scorer names", () => {
    const report = runEval(evalSetPath(), demoTarget);
    expect(report.cases.map((c) => c.weight)).toEqual([2.0, 1.0, 1.5]);
    expect(report.cases.map((c) => c.scorer)).toEqual(["contains", "semantic", "regex"]);
  });

  it("lets weights shift the aggregate while staying in range", () => {
    const report = runEval(evalSetPath(), demoTarget);
    expect(report.cases.some((c) => c.weight === 2.0)).toBe(true);
    expect(report.aggregate).toBeGreaterThanOrEqual(0.0);
    expect(report.aggregate).toBeLessThanOrEqual(1.0);
  });

  it("scores and passes each individual case", () => {
    const report = runEval(evalSetPath(), demoTarget);
    expect(report.cases[0].score).toBe(1.0);
    expect(report.cases[0].passed).toBe(true);
    expect(report.cases[1].score).toBeCloseTo(0.6573, 10);
    expect(report.cases[1].passed).toBe(true);
    expect(report.cases[2].score).toBe(1.0);
    expect(report.cases[2].passed).toBe(true);
  });

  it("fails the gate for a regressed target", () => {
    const regressed = (q: string) =>
      q.toLowerCase().includes("cancel") ? "Yes, cancel anytime." : "I don't know.";
    const report = runEval(evalSetPath(), regressed);
    expect(report.passed).toBe(false);
  });

  it("lets an explicit threshold override a higher bar", () => {
    const report = runEval(evalSetPath(), demoTarget, 0.99);
    expect(report.passed).toBe(false);
    expect(report.threshold).toBe(0.99);
  });

  it("lets an explicit threshold override a lower bar", () => {
    const report = runEval(evalSetPath(), demoTarget, 0.1);
    expect(report.passed).toBe(true);
    expect(report.threshold).toBe(0.1);
  });

  it("throws for a missing eval set file", () => {
    expect(() => runEval("does-not-exist.yaml", demoTarget)).toThrow();
  });

  it("throws for an unknown scorer", () => {
    const path = writeTempEval(
      "threshold: 0.5\ncases:\n  - id: c1\n    input: hello\n    scorer: bogus\n    expected: hello\n",
    );
    expect(() => runEval(path, demoTarget)).toThrow(/unknown scorer/);
  });

  it("falls back to weight 1.0 when all weights are zero", () => {
    const path = writeTempEval(
      "threshold: 0.0\ncases:\n  - id: a\n    input: x\n    scorer: exact_match\n    expected: x\n    weight: 0.0\n  - id: b\n    input: y\n    scorer: exact_match\n    expected: y\n    weight: 0.0\n",
    );
    const report = runEval(path, demoTarget);
    expect(report.aggregate).toBe(0.0);
    expect(report.passed).toBe(true);
  });

  it("defaults the threshold to 0.8 when the eval set omits it", () => {
    const path = writeTempEval(
      "cases:\n  - id: a\n    input: x\n    scorer: exact_match\n    expected: x\n",
    );
    const echo = (s: string) => s;
    const report = runEval(path, echo);
    expect(report.threshold).toBe(0.8);
    expect(report.aggregate).toBe(1.0);
  });
});

describe("demoTarget", () => {
  it("answers the refund branch", () => {
    expect(demoTarget("How many days for a refund?")).toBe(
      "You have 30 days from purchase to request a full refund.",
    );
  });
  it("answers the cancel branch", () => {
    expect(demoTarget("Can I cancel?")).toBe(
      "Yes, you can cancel your subscription at any time; it ends at the cycle close.",
    );
  });
  it("answers the discount branch", () => {
    expect(demoTarget("Any discount?")).toBe("No, we do not offer a 90% loyalty discount.");
  });
  it("declines everything else", () => {
    expect(demoTarget("What is the weather?")).toBe("I don't have information on that.");
  });
});
