/** EvalForge scorers: turn a target output and an argument into a 0..1 score. */

/**
 * A scorer turns a target output and a single argument (the expected string, or
 * the pattern for the regex scorer) into a score in the range 0..1.
 */
export type Scorer = (output: string, arg: string) => number;

/**
 * Scores 1.0 when output and expected are equal after trimming surrounding
 * whitespace, else 0.0.
 */
export function exactMatch(output: string, expected: string): number {
  return output.trim() === expected.trim() ? 1.0 : 0.0;
}

/**
 * Scores 1.0 when expected (trimmed and lower-cased) appears anywhere in output
 * (lower-cased), else 0.0.
 */
export function contains(output: string, expected: string): number {
  return output.toLowerCase().includes(expected.trim().toLowerCase()) ? 1.0 : 0.0;
}

/**
 * Scores 1.0 when pattern matches anywhere in output, else 0.0. A leading inline
 * flag group such as `(?i)` is translated to JavaScript RegExp flags, because V8
 * does not accept the bare `(?i)` prefix that Python, RE2 and the Rust regex
 * crate allow. An invalid pattern throws, mirroring the Python reference raising
 * `re.error`.
 */
export function regexMatch(output: string, pattern: string): number {
  return compileRegex(pattern).test(output) ? 1.0 : 0.0;
}

function compileRegex(pattern: string): RegExp {
  const m = /^\(\?([a-zA-Z]+)\)/.exec(pattern);
  if (m) {
    let flags = "";
    for (const ch of m[1]) {
      // Python's (?aiLmsux) prefix flags; only i/m/s have direct JS RegExp
      // equivalents, and the eval sets only use i.
      if (ch === "i" && !flags.includes("i")) flags += "i";
      else if (ch === "m" && !flags.includes("m")) flags += "m";
      else if (ch === "s" && !flags.includes("s")) flags += "s";
    }
    return new RegExp(pattern.slice(m[0].length), flags);
  }
  return new RegExp(pattern);
}

/**
 * A lightweight, dependency-free proxy for semantic closeness: the mean of
 * Jaccard token overlap and difflib's sequence ratio, rounded to four decimals.
 * Not a real embedding model; the scorer interface is identical so it can be
 * swapped for one in production.
 */
export function semanticSimilarity(output: string, expected: string): number {
  const a = output
    .toLowerCase()
    .split(/\s+/)
    .filter((w) => w.length > 0);
  const b = expected
    .toLowerCase()
    .split(/\s+/)
    .filter((w) => w.length > 0);
  if (a.length === 0 || b.length === 0) return 0.0;
  const sa = new Set(a);
  const sb = new Set(b);
  let inter = 0;
  for (const w of sa) if (sb.has(w)) inter++;
  const union = new Set([...sa, ...sb]).size;
  const overlap = inter / union;
  const ratio = sequenceRatio(output.toLowerCase(), expected.toLowerCase());
  return round4(0.5 * overlap + 0.5 * ratio);
}

/**
 * Rounds v to four decimal places, reproducing the value Python's `round(v, 4)`
 * yields for the same input. The values scored here never fall exactly on a
 * 4th-decimal tie, so the tie-breaking difference between `toFixed` and Python's
 * round-half-to-even does not arise.
 */
export function round4(v: number): number {
  return Number(v.toFixed(4));
}

/**
 * Reproduces `difflib.SequenceMatcher(null, a, b).ratio()`: it is `2*M/T` where
 * `T` is the combined length of both strings and `M` is the total size of the
 * matching blocks found by the Ratcliff/Obershelp algorithm. Comparison is per
 * Unicode code point, and the autojunk heuristic is not applied (it only affects
 * sequences of 200+ elements, well beyond the short strings scored here).
 */
function sequenceRatio(a: string, b: string): number {
  const ra = Array.from(a);
  const rb = Array.from(b);
  if (ra.length + rb.length === 0) return 1.0;
  const b2j = new Map<string, number[]>();
  rb.forEach((ch, j) => {
    const arr = b2j.get(ch);
    if (arr) arr.push(j);
    else b2j.set(ch, [j]);
  });
  const matches = totalMatches(ra, b2j, 0, ra.length, 0, rb.length);
  return (2.0 * matches) / (ra.length + rb.length);
}

function totalMatches(
  a: string[],
  b2j: Map<string, number[]>,
  alo: number,
  ahi: number,
  blo: number,
  bhi: number,
): number {
  const [i, j, k] = findLongestMatch(a, b2j, alo, ahi, blo, bhi);
  if (k === 0) return 0;
  return (
    k +
    totalMatches(a, b2j, alo, i, blo, j) +
    totalMatches(a, b2j, i + k, ahi, j + k, bhi)
  );
}

function findLongestMatch(
  a: string[],
  b2j: Map<string, number[]>,
  alo: number,
  ahi: number,
  blo: number,
  bhi: number,
): [number, number, number] {
  let besti = alo;
  let bestj = blo;
  let bestsize = 0;
  let j2len = new Map<number, number>();
  for (let i = alo; i < ahi; i++) {
    const newj2len = new Map<number, number>();
    const js = b2j.get(a[i]);
    if (js) {
      for (const j of js) {
        if (j < blo) continue;
        if (j >= bhi) break;
        const k = (j2len.get(j - 1) ?? 0) + 1;
        newj2len.set(j, k);
        if (k > bestsize) {
          besti = i - k + 1;
          bestj = j - k + 1;
          bestsize = k;
        }
      }
    }
    j2len = newj2len;
  }
  return [besti, bestj, bestsize];
}

const SCORERS: Record<string, Scorer> = {
  exact_match: exactMatch,
  contains,
  regex: regexMatch,
  semantic: semanticSimilarity,
};

/** The built-in scorer names, in registration order. */
export const SCORER_NAMES: readonly string[] = ["exact_match", "contains", "regex", "semantic"];

/** Returns the names of the built-in scorers. */
export function scorerNames(): string[] {
  return [...SCORER_NAMES];
}

/** Reports whether name is a built-in scorer. */
export function isKnownScorer(name: string): boolean {
  return SCORER_NAMES.includes(name);
}

/**
 * Returns the scorer registered under name, or throws if no such scorer exists,
 * mirroring the Python reference's `get_scorer`.
 */
export function getScorer(name: string): Scorer {
  if (!Object.prototype.hasOwnProperty.call(SCORERS, name)) {
    throw new Error(`unknown scorer '${name}'. available: ${JSON.stringify(scorerNames())}`);
  }
  return SCORERS[name];
}
