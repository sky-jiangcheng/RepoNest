package abeval

import (
	"math"
	"testing"
)

func approx(t *testing.T, got, want float64, msg string) {
	t.Helper()
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("%s: got %.6f, want %.6f", msg, got, want)
	}
}

func TestRecallAtK(t *testing.T) {
	approx(t, RecallAtK([]int64{1, 2, 3}, []int64{2, 5}, 2), 0.5, "half relevant in top-2")
	approx(t, RecallAtK([]int64{1, 1, 2}, []int64{2}, 2), 0, "dup at pos0; relevant at pos2 outside window")
	approx(t, RecallAtK([]int64{5, 6, 7}, []int64{6, 7}, 3), 1.0, "all relevant within k")
	approx(t, RecallAtK([]int64{1}, nil, 3), 0, "no relevant -> 0")
	approx(t, RecallAtK([]int64{1}, []int64{1}, 0), 0, "k<=0 -> 0")
}

func TestNDCGAtK(t *testing.T) {
	approx(t, NDCGAtK([]int64{2, 3}, []int64{2, 3}, 3), 1.0, "perfect ranking")
	// 2@pos0 (gain 1.0) + 3@pos2 (gain 1/log2(4)=0.5) over IDCG (1 + 1/log2(3)).
	approx(t, NDCGAtK([]int64{2, 1, 3}, []int64{2, 3}, 3), 1.5/(1+1/math.Log2(3)), "relevant at 0 and 2")
	approx(t, NDCGAtK([]int64{1, 2}, nil, 2), 0, "no relevant -> 0")
}

func TestScoreSkipsUnlabeled(t *testing.T) {
	m := Score([]Retrieved{
		{Case: Case{"", nil}, Ranks: []int64{1}},
		{Case: Case{"", []int64{2, 5}}, Ranks: []int64{2, 3}},
	}, 3)
	if m.N != 1 {
		t.Fatalf("N=%d, want 1 (unlabeled skipped)", m.N)
	}
	approx(t, m.MeanRecall, 0.5, "1 of 2 relevant in top-3")
}

func TestCompareHybridBeatsLexical(t *testing.T) {
	cases := []Case{
		{Query: "q1", Relevant: []int64{10}},
		{Query: "q2", Relevant: []int64{20, 21}},
	}
	lexical := func(q string) []int64 {
		if q == "q1" {
			return []int64{99} // misses 10
		}
		return []int64{20} // gets one of two
	}
	hybrid := func(q string) []int64 {
		if q == "q1" {
			return []int64{10} // finds it
		}
		return []int64{21, 20} // finds both
	}
	lex, hyb, dRecall, dNDCG := Compare(cases, 3, lexical, hybrid)
	if !(hyb.MeanRecall > lex.MeanRecall) {
		t.Fatalf("expected hybrid recall > lexical; lex=%.3f hyb=%.3f", lex.MeanRecall, hyb.MeanRecall)
	}
	if dRecall <= 0 || dNDCG <= 0 {
		t.Fatalf("expected positive deltas; dRecall=%.3f dNDCG=%.3f", dRecall, dNDCG)
	}
}

func TestCompareNilRetrieverSafe(t *testing.T) {
	lex, hyb, _, _ := Compare([]Case{{"x", []int64{1}}}, 3, nil, func(string) []int64 { return []int64{1} })
	approx(t, lex.MeanRecall, 0, "nil lexical -> 0")
	approx(t, hyb.MeanRecall, 1, "hybrid finds it")
}

func TestSummaryLine(t *testing.T) {
	line := Metrics{K: 5, N: 2, MeanRecall: 0.5, MeanNDCG: 0.25}.SummaryLine("hybrid")
	if line == "" {
		t.Fatal("empty line")
	}
}
