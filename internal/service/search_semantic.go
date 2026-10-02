package service

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"repo-nest/internal/db"
	"repo-nest/internal/domain"
	"repo-nest/internal/search/hybrid"
)

// M3-A semantic search wiring (ADR-0012). Default OFF: nothing here runs unless
// `semantic_search=="1"`. Even when enabled, every failure (endpoint not set,
// embed call error, vector index absent) degrades gracefully to the existing
// lexical result — so enabling it can never make search return less than before.

const (
	semanticConfigKey = "semantic_search"
	embedBatch        = 64 // texts per embedding request
	embedRecallK      = 20 // vector ids fused with the lexical hits
)

// semanticEnabled reports whether the (opt-in) semantic search is on.
func (s *Service) semanticEnabled() bool {
	v, err := db.GetConfig(s.db, semanticConfigKey)
	return err == nil && v == "1"
}

// noteEmbedder builds a RemoteEmbedder from config; ok=false when base_url or
// model is unset. dim may be 0 ("learn it from the first embedding").
func (s *Service) noteEmbedder() (*hybrid.RemoteEmbedder, bool) {
	base, _ := db.GetConfig(s.db, "embedding_base_url")
	model, _ := db.GetConfig(s.db, "embedding_model")
	if strings.TrimSpace(base) == "" || strings.TrimSpace(model) == "" {
		return nil, false
	}
	key, _ := db.GetConfig(s.db, "embedding_api_key")
	dimStr, _ := db.GetConfig(s.db, "embedding_dim")
	dim, _ := strconv.Atoi(dimStr)
	return &hybrid.RemoteEmbedder{BaseURL: base, Model: model, APIKey: key, Dim: dim}, true
}

// RebuildEmbeddings fully (re)embeds every note into the vector index. Used when
// the user turns on semantic search or changes the endpoint/model. Returns the
// number of notes embedded. If embedding_dim is unset it is learned from the
// first embedding. The API key is read here (never returned to the UI).
func (s *Service) RebuildEmbeddings() (int, error) {
	emb, ok := s.noteEmbedder()
	if !ok {
		return 0, fmt.Errorf("semantic search: embedding endpoint not configured")
	}
	notes, err := db.ListNoteEmbeddingInputs(s.db)
	if err != nil {
		return 0, err
	}
	if emb.Dim <= 0 {
		if len(notes) == 0 {
			return 0, fmt.Errorf("semantic search: no notes to infer embedding dim from; set embedding_dim")
		}
		probe, perr := emb.Embed([]string{notes[0].Text})
		if perr != nil {
			return 0, perr
		}
		if len(probe) == 0 || len(probe[0]) == 0 {
			return 0, fmt.Errorf("semantic search: endpoint returned an empty vector")
		}
		emb.Dim = len(probe[0])
	}
	if err := db.EnsureVectorIndex(s.db, emb.Dim); err != nil {
		return 0, err
	}
	if err := db.ClearVectorIndex(s.db); err != nil {
		return 0, err
	}
	embedded := 0
	for start := 0; start < len(notes); start += embedBatch {
		end := start + embedBatch
		if end > len(notes) {
			end = len(notes)
		}
		texts := make([]string, 0, end-start)
		ids := make([]int64, 0, end-start)
		for _, n := range notes[start:end] {
			texts = append(texts, n.Text)
			ids = append(ids, n.ID)
		}
		vecs, err := emb.Embed(texts)
		if err != nil {
			log.Printf("semantic rebuild: embed batch %d: %v", start, err)
			break
		}
		for i, vec := range vecs {
			if i < len(ids) {
				if err := db.PutNoteEmbedding(s.db, ids[i], vec); err != nil {
					return embedded, err
				}
				embedded++
			}
		}
	}
	return embedded, nil
}

// fuseSemantic merges vector recall into the lexical hits via RRF. Returns the
// input unchanged when semantic search is off, unconfigured, the index is
// missing, or any embed/KNN step fails — never reducing results.
func (s *Service) fuseSemantic(base []domain.SearchHit, query string) []domain.SearchHit {
	if !s.semanticEnabled() || !db.VectorIndexReady(s.db) {
		return base
	}
	emb, ok := s.noteEmbedder()
	if !ok {
		return base
	}
	vecs, err := emb.Embed([]string{query})
	if err != nil || len(vecs) == 0 {
		if err != nil {
			log.Printf("semantic query embed failed; using lexical only: %v", err)
		}
		return base
	}
	ids, err := db.KnnNoteIDs(s.db, vecs[0], embedRecallK)
	if err != nil || len(ids) == 0 {
		return base
	}
	baseIDs := make(hybrid.RankList, 0, len(base))
	byID := make(map[int64]domain.SearchHit, len(base))
	for _, h := range base {
		baseIDs = append(baseIDs, h.ID)
		byID[h.ID] = h
	}
	merged := hybrid.FuseRRF(0, baseIDs, hybrid.RankList(ids))

	// Fetch any vector-only ids we don't already hold.
	var missing []int64
	for _, id := range merged {
		if _, have := byID[id]; !have {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		extra, err := db.NoteHitsByIDs(s.db, missing, query)
		if err == nil {
			for _, h := range extra {
				byID[h.ID] = h
			}
		}
	}
	out := make([]domain.SearchHit, 0, len(merged))
	for _, id := range merged {
		if h, have := byID[id]; have {
			out = append(out, h)
		}
	}
	return out
}
