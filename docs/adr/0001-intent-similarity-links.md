# ADR-0001: IDF-Weighted Cosine Similarity for Intent Links

- **Status:** Accepted
- **Date:** 2026-07-31
- **Owner:** graph package (`internal/graph/similarity.go`)

## Context

`dizz` links intents and TODOs with `RELATED_TO` edges so the graph and the
web visualization can show "these project-level records are about the same
topic". The original implementation scored every pair with the **overlap
coefficient** over raw bag-of-words token sets:

```
score(A, B) = |tokens(A) ∩ tokens(B)| / min(|tokens(A)|, |tokens(B)|)
```

Empirically this was wrong in two opposite directions:

1. **False positives.** `"test intent"` scored 0.5 against
   `"Add description for intents and command intent describe..."` because the
   short message shares one generic token (`intent`) with a long one. The
   overlap coefficient is biased toward short texts: a 2-token message matching
   any 2 tokens of a 10-token message scores 1.0.
2. **False negatives.** `"Fix cache invalidation race condition"` scored 0
   against `"Should we add caching for analysis results?"` - both about caching,
   but `caching` and `cache` are different tokens and the raw coefficient has no
   inflection handling and no notion of how distinctive a token is.

## Decision

Replace the overlap coefficient with **cosine similarity over IDF-weighted,
stemmed token vectors**, computed with an **inverted index** so only pairs that
share at least one token are ever scored. The complete algorithm is specified
below. The full set of changes lands in `internal/graph/similarity.go`; the
`--similarity-threshold` and `--similarity-topk` CLI flags are unchanged in
shape, only their defaults move (see Configuration).

## Algorithm

### 1. Candidate pool

A candidate is one project-level record: every intent (regardless of status,
including resolved and closed) plus every unresolved TODO. Intents and TODOs
only ever match each other - never symbols, files, or other node types.

The compared text for each candidate is `Message + tags`, with each tag
repeated once so tags carry double weight:

```
text = msg + " " + tag1 + " " + tag1 + " " + tag2 + " " + tag2 + ...
```

Scope and file paths are deliberately excluded: intent scopes are dominated by
constant values (e.g. `project`) and path tokens (`pkg`, `go`) that only add
noise. TODO file context is already represented by the graph's `HAS_TODO` edges.

### 2. Tokenization

For each candidate text, in order:

1. **Lowercase** the text.
2. **Split** on every non-alphanumeric run (only `[a-z0-9]` is kept).
3. **Drop stopwords**: a fixed set of function words with no topical meaning
   (`the`, `to`, `of`, `for`, `and`, `should`, `this`, ...). Action verbs
   (`refactor`, `fix`, `add`) are NOT stopwords - they carry meaning, and IDF
   downweights them naturally when they appear everywhere.
4. **Drop single-character** tokens.
5. **Stem** each remaining token with a conservative suffix stripper:

```
- "ies"            -> "y"          (abilities -> abilit)
- "sses"           -> "ss"         (classes  -> class)
- "es"             -> ""           (caches   -> cache)
- "s"              -> ""           (tokens   -> token)   [not ss/us/is]
- "ing" (len>5)    -> ""           (caching  -> cach, refactoring -> refactor)
- "ed"  (len>4)    -> ""           (resolved -> resolv)
- "ly"  (len>4)    -> ""           (quickly  -> quick)
- trailing "e"     -> ""           (cache    -> cach)    [not ee/ie]
```

The trailing-`e` rule is what makes `caching` (-> `cach`) match `cache`
(-> `cach`). The stemmer is deliberately conservative: uncommon or ambiguous
suffixes are left untouched rather than risking false matches. Duplicates are
preserved (term frequency matters).

### 3. IDF weighting

For the whole corpus of `N` candidates:

```
df(t)  = number of candidates containing stem t
idf(t) = ln(1 + N / df(t))
w(t)   = tf(t) * idf(t)     where tf(t) is the count of t in this candidate
```

IDF is the whole point: a token shared by every candidate (`intent` in this
repository) is weighted far below a token appearing once (`invalidation`).
Generic words therefore cannot dominate a score, which fixes the false-positive
class of bug directly.

### 4. Cosine similarity

Each candidate `i` becomes a sparse vector `V_i` of term weights plus its
precomputed L2 norm:

```
norm(i) = sqrt( Σ_t w_i(t)^2 )

score(i, j) = ( Σ_t w_i(t) * w_j(t) ) / ( norm(i) * norm(j) )
```

Scores are in `[0, 1]`; `0` means no shared stem, `1` means identical weighted
term sets.

### 5. Pair selection (inverted index + accumulators)

Exact all-pairs scoring is worst-case O(N^2), so it is never done. Instead a
term postings index is built (`term -> sorted list of candidate indices`) and
for each candidate `i`:

```
acc = {}                                # candidate index -> partial dot product
for term t in sorted(terms of i):       # sorted for determinism
    for j in postings[t]:
        if j != i: acc[j] += w_i(t) * w_j(t)

for j, dot in acc:
    sim = dot / (norm(i) * norm(j))
    keep if sim >= threshold
sort kept pairs by sim desc (ties by node id), truncate to topK
emit a RELATED_TO edge in each direction
```

Only pairs sharing at least one stem are ever scored, so the practical cost is
O(Σ_t |postings(t)|^2), which is near-linear for sparse vocabularies and only
degrades to O(N^2) when every document shares a common term. A provable O(N)
result would require approximation (MinHash/LSH) or single-query search; with
the intended corpus (tens of intents) exact scoring is microseconds and
approximation is not justified. This design keeps the exact answer while
retaining a scalable pattern if the corpus ever grows.

### 6. Determinism

The whole pipeline is deterministic:

- Tokenization, stemming, stopwords, and IDF are pure functions.
- Postings lists are built in ascending candidate order.
- Terms are iterated in sorted order, so the floating-point dot product is
  accumulated in a fixed order.
- Ties in the topK sort are broken by node id.

Two identical builds always produce identical edges and weights.

### 7. Edge shape

Each selected pair produces a `RELATED_TO` edge in both directions (deduplicated
by `EdgeKey`), carrying:

- `Weight` = the cosine score (used by the web view for link width and the
  similarity % tooltip).
- `attrs["similarity"]` = the score formatted to 3 decimals.
- `Rationale` = `{Confidence: score, Evidence: "text similarity 0.230", SourceType: "similarity"}`.

## Configuration

| Flag | Default | Meaning |
|------|---------|---------|
| `--similarity-threshold` | `0.2` | Minimum cosine score for a `RELATED_TO` edge |
| `--similarity-topk` | `6` | Max `RELATED_TO` edges per intent |

Defaults live in `DefaultBuildOptions` (`internal/graph/builder.go`) and the
flag definitions in `cmd/graph.go` and `cmd/visualize.go`.

**Calibration** (this repository, 10 intents + 1 TODO, no scope in the compared
text):

| Score | Pair | Verdict |
|-------|------|---------|
| 0.230 | resolveTime-to-intents <-> scope-determination-logic-in-intents | genuine (intent tooling) - linked |
| 0.183 | add-description-for-intents <-> test intent | junk false positive - excluded |
| 0.144 | fix-cache-invalidation <-> add-caching | genuine (caching) - below default, reachable with a lower threshold |
| 0.172 (fixture) | refactor-auth <-> refactor-token-system | single shared generic stem - excluded |

A threshold of 0.2 keeps strong single-topic-overlap pairs while excluding
weak one-generic-token coincidences. Users wanting aggressive recall (e.g. the
0.144 caching pair) can lower the threshold.

## Alternatives considered

- **Raw overlap coefficient (status quo).** Rejected: length-bias toward short
  messages causes the false-positive class of bug.
- **MinHash + LSH.** Rejected: approximate, and only pays off at 10^5+
  documents. The exact inverted-index accumulator is faster to the point of
  being free at the intended scale.
- **SimHash.** Rejected: 64-bit fingerprints are too coarse to rank short
  sentences.
- **Character n-gram shingles.** Rejected for the primary metric: handles
  inflections and typos without a stemmer, but dilutes scores on longer texts
  (a 40-gram document barely moves even on a genuine match). Documented here as
  a possible fallback signal if recall needs to improve.
- **Full Porter stemmer.** Rejected: ~200 lines and notoriously subtle edge
  cases; the 10-line suffix stripper covers the inflections that actually occur
  in intent messages.
- **Embeddings / LLM.** Rejected: external dependencies, non-deterministic, and
  offline-hostile. dizz must work with zero configuration and zero new deps.
- **Storing embeddings / re-scoring at query time.** Rejected: edges are
  derived at graph-build time and snapshotted; determinism is a hard
  requirement.

## Consequences

**Positive:**

- False positives from shared generic words are eliminated (demonstrated: the
  `test intent` link dropped from 0.5 to 0.183 and is excluded at the default).
- Inflection variants now match (`cache`/`caching`, `refactor`/`refactoring`,
  `token`/`tokens`) - the caching pair that scored 0 now scores 0.144.
- Zero new dependencies; fully deterministic; exact (no approximation).
- Scaling pattern (inverted index) is in place if the corpus grows.

**Negative:**

- Scores are corpus-relative (IDF depends on the candidate pool), so the same
  pair scores differently in different repositories; the threshold default is
  calibrated, not proven.
- One-generic-word overlap pairs (the top genuine link in this repo at 0.230)
  are indistinguishable in kind from junk pairs (0.183) - only the score
  separates them, and the gap is data-dependent.
- Missed paraphrases that share no stems (e.g. `description` vs `describe`) -
  the stemmer is intentionally conservative.
