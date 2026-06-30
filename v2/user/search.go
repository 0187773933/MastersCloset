package user

import (
	"strings"

	bleve "github.com/blevesearch/bleve/v2"
	bolt "github.com/boltdb/bolt"
)

// SearchExact resolves a user UUID from a display name (spaces or hyphens) via
// the username index. Returns "" if not found.
func (s *Store) SearchExact(name string) string {
	formatted := strings.ReplaceAll(strings.TrimSpace(name), " ", "-")
	var id string
	s.db.View(func(tx *bolt.Tx) error {
		if v := tx.Bucket([]byte(bucketUsernames)).Get([]byte(formatted)); v != nil {
			id = string(v)
		}
		return nil
	})
	return id
}

// SearchFuzzy powers the live type-ahead search. For each typed word it matches
// either a name token that starts with it (prefix — so "pear" finds "Pearson"
// as you type) OR a token within edit-distance 1 (fuzzy — tolerates typos like
// "pearsen"). All words must match. Index entries whose user was deleted are
// skipped.
func (s *Store) SearchFuzzy(query string) []User {
	if s.index == nil {
		return nil
	}
	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 {
		return nil
	}
	boolQuery := bleve.NewBooleanQuery()
	for _, w := range words {
		prefix := bleve.NewPrefixQuery(w)
		prefix.SetField("Name")
		fuzzy := bleve.NewFuzzyQuery(w)
		fuzzy.Fuzziness = 1
		fuzzy.SetField("Name")
		boolQuery.AddMust(bleve.NewDisjunctionQuery(prefix, fuzzy))
	}
	req := bleve.NewSearchRequest(boolQuery)
	req.Size = 25
	results, err := s.index.Search(req)
	if err != nil {
		return nil
	}
	var out []User
	for _, hit := range results.Hits {
		if u, ok := s.Get(hit.ID); ok {
			out = append(out, u)
		}
	}
	return out
}
