/** EvalForge runner: score an eval set against a target and gate on a threshold. */

import { readFileSync } from "node:fs";
import { parse } from "yaml";
import { getScorer, round4 } from "./scorers";

/** The scored outcome of a single eval case. */
export interface CaseResult {
  /** The case identifier from the eval set. */
  id: string;
  /** The name of the scorer that graded the case. */
  scorer: string;
  /** The case's contribution to the weighted aggregate. */
  weight: number;
  /** The scorer's 0..1 result for the target output. */
  score: number;
  /** Whether score met the case's pass_score threshold. */
  passed: boolean;
}

/** The aggregate outcome of running an eval set. */
export interface EvalReport {
  /** The weighted mean case score, rounded to four decimals. */
  aggregate: number;
  /** The gate the aggregate had to meet to pass. */
  threshold: number;
  /** Whether aggregate met threshold. */
  passed: boolean;
  /** The per-case results in eval-set order. */
  cases: CaseResult[];
}

/** A target under evaluation: it maps a question to an answer. */
export type Target = (question: string) => string;

interface CaseSpec {
  id: string;
  input: string;
  scorer: string;
  expected?: string;
  pattern?: string;
  weight?: number;
  pass_score?: number;
}

interface EvalSpec {
  threshold?: number;
  cases?: CaseSpec[];
}

/**
 * Loads the eval set at evalSetPath, scores each case against target, and gates
 * the weighted-mean aggregate on a threshold. A defined threshold overrides the
 * eval set's own threshold (which defaults to 0.8 when absent). This is the
 * "unit tests for prompts" engine: in CI a failing gate is meant to block the
 * deploy.
 */
export function runEval(evalSetPath: string, target: Target, threshold?: number): EvalReport {
  const data = readFileSync(evalSetPath, "utf8");
  const spec = (parse(data) ?? {}) as EvalSpec;

  let effective = spec.threshold ?? 0.8;
  if (threshold !== undefined) {
    effective = threshold;
  }

  const cases: CaseResult[] = [];
  for (const c of spec.cases ?? []) {
    const output = target(c.input);
    const scorer = getScorer(c.scorer);
    const arg = c.scorer === "regex" ? (c.pattern ?? "") : (c.expected ?? "");
    const score = scorer(output, arg);
    const weight = c.weight ?? 1.0;
    const caseThreshold = c.pass_score ?? 0.5;
    cases.push({
      id: c.id,
      scorer: c.scorer,
      weight,
      score,
      passed: score >= caseThreshold,
    });
  }

  let totalW = cases.reduce((sum, r) => sum + r.weight, 0);
  if (totalW === 0) totalW = 1.0;
  const weighted = cases.reduce((sum, r) => sum + r.score * r.weight, 0);
  const aggregate = round4(weighted / totalW);

  return {
    aggregate,
    threshold: effective,
    passed: aggregate >= effective,
    cases,
  };
}
