// Package abeval is the M3-A quality gate (ADR-0012 决策 4): an offline harness
// that measures whether semantic (hybrid) search actually beats lexical
// (FTS5+relaxation) search on a labeled query set — before it is ever exposed to
// users. It is pure computation over ranked id lists, so it has no DB, network,
// or embedding dependencies: the caller runs the two retrievers (e.g.
// SearchNotes with semantic_search on vs off) and feeds the results in.
//
// Metrics are binary-relevance (a doc is relevant or not), the honest bar for
// "does this retrieval surface the docs it should" on a personal corpus.
package abeval

import (
	"fmt"
	"math"
)

// Case is one labeled query: what the user typed and the set of note ids that
// actually answer it. A case with no relevant ids is skipped by the scorers.
type Case struct {
	Query    string
	Relevant []int64
}

// Retrieved is a ranked result (best first) for one case.
type Retrieved struct {
	Case  Case
	Ranks []int64
}

// Metrics holds aggregate quality numbers over retrievals at cutoff k.
type Metrics struct {
	K          int
	N          int     // cases scored (non-empty Relevant)
	MeanRecall float64 // mean |topK ∩ relevant| / |relevant|
	MeanNDCG   float64 // mean normalized DCG@K (binary relevance)
}

// RecallAtK = |ranks[:k] ∩ relevant| / |relevant|; 0 when relevant is empty.
// Repeated ids in ranks are counted once.
func RecallAtK(ranks, relevant []int64, k int) float64 {
	if len(relevant) == 0 || k <= 0 {
		return 0
	}
	set := idSet(relevant)
	seen := make(map[int64]struct{}, k)
	hits := 0
	for i, id := range ranks {
		if i >= k {
			break
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		if _, ok := set[id]; ok {
			hits++
		}
	}
	return float64(hits) / float64(len(set))
}

// NDCGAtK is binary-relevance normalized DCG@K. Repeated ids count once (they
// occupy their first position only), matching how a user sees a ranked list.
func NDCGAtK(ranks, relevant []int64, k int) float64 {
	if len(relevant) == 0 || k <= 0 {
		return 0
	}
	set := idSet(relevant)
	seen := make(map[int64]struct{}, k)
	dcg := 0.0
	for i, id := range ranks {
		if i >= k {
			break
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		if _, ok := set[id]; ok {
			dcg += 1.0 / math.Log2(float64(i)+2.0) // position i (0-based) -> log2(i+2)
		}
	}
	idcg := 0.0
	for i := 0; i < k && i < len(set); i++ {
		idcg += 1.0 / math.Log2(float64(i)+2.0)
	}
	if idcg == 0 {
		return 0
	}
	return dcg / idcg
}

// Score aggregates Recall@k and NDCG@k over retrievals, skipping unlabeled cases.
func Score(retrievals []Retrieved, k int) Metrics {
	m := Metrics{K: k}
	if k <= 0 {
		return m
	}
	var sumR, sumN float64
	for _, r := range retrievals {
		if len(r.Case.Relevant) == 0 {
			continue
		}
		sumR += RecallAtK(r.Ranks, r.Case.Relevant, k)
		sumN += NDCGAtK(r.Ranks, r.Case.Relevant, k)
		m.N++
	}
	if m.N > 0 {
		m.MeanRecall = sumR / float64(m.N)
		m.MeanNDCG = sumN / float64(m.N)
	}
	return m
}

// Compare runs the same labeled cases through two retrievers (lexical and
// hybrid) and returns both metric sets plus the deltas (hybrid - lexical). A
// positive delta over a real labeled set is the evidence ADR-0012 requires
// before enabling semantic search for users. A nil retriever yields empty ranks.
func Compare(cases []Case, k int, lexical, hybrid func(query string) []int64) (lex, hyb Metrics, dRecall, dNDCG float64) {
	lexRet := make([]Retrieved, 0, len(cases))
	hybRet := make([]Retrieved, 0, len(cases))
	for _, c := range cases {
		lexRet = append(lexRet, Retrieved{Case: c, Ranks: call(lexical, c.Query)})
		hybRet = append(hybRet, Retrieved{Case: c, Ranks: call(hybrid, c.Query)})
	}
	lex = Score(lexRet, k)
	hyb = Score(hybRet, k)
	return lex, hyb, hyb.MeanRecall - lex.MeanRecall, hyb.MeanNDCG - lex.MeanNDCG
}

// SummaryLine is a compact one-line report for logging / eval output.
func (m Metrics) SummaryLine(label string) string {
	return fmt.Sprintf("%s: n=%d recall@%d=%.3f ndcg@%d=%.3f", label, m.N, m.K, m.MeanRecall, m.K, m.MeanNDCG)
}

func call(f func(string) []int64, q string) []int64 {
	if f == nil {
		return nil
	}
	return f(q)
}

func idSet(ids []int64) map[int64]struct{} {
	m := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		m[id] = struct{}{}
	}
	return m
}
