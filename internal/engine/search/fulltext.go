package search

import (
	"context"
	"fmt"
	"strings"
	"unicode"
)

type FullTextIndex struct {
	inverted map[string]map[string][]string // term -> objType -> [objIDs]
}

func NewFullTextIndex() *FullTextIndex {
	return &FullTextIndex{
		inverted: make(map[string]map[string][]string),
	}
}

func (idx *FullTextIndex) Index(ctx context.Context, objType, objID string, fields map[string]any) error {
	for _, value := range fields {
		terms := tokenize(fmt.Sprintf("%v", value))
		for _, term := range terms {
			if idx.inverted[term] == nil {
				idx.inverted[term] = make(map[string][]string)
			}
			idx.inverted[term][objType] = append(idx.inverted[term][objType], objID)
		}
	}
	return nil
}

func (idx *FullTextIndex) Search(ctx context.Context, query string) ([]string, error) {
	terms := tokenize(query)
	if len(terms) == 0 {
		return nil, nil
	}

	resultSet := make(map[string]bool)
	for _, term := range terms {
		term = strings.ToLower(term)
		if posting, ok := idx.inverted[term]; ok {
			for _, ids := range posting {
				for _, id := range ids {
					resultSet[id] = true
				}
			}
		}
	}

	var results []string
	for id := range resultSet {
		results = append(results, id)
	}
	return results, nil
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}