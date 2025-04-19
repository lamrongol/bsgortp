package bsgortp

import (
	"testing"
)

func TestEmptyTextPost(t *testing.T) {
	_, err := GenPost("", []string{"en"})
	if err == nil {
		t.Errorf("empty string should cause an error")
	}
}

func TestSimplePosts(t *testing.T) {
	tests := []string{
		"hi!",
		"I'm in that mode?",
	}

	for _, tt := range tests {
		post, err := GenPost(tt, []string{"en"})

		if err != nil {
			t.Errorf("input=%s failed: %s", tt, err)
		}

		if len(post.Facets) != 0 {
			t.Errorf("expected 0 facets from input=%s", tt)
		}
	}
}

func TestPostsWithLinks(t *testing.T) {
	tests := []struct {
		Input              string
		ExpectedUrl        string
		ExpectedFacetCount int
		ExpectedByteStart  int64
		ExpectedByteEnd    int64
	}{
		{
			"go visit https://cats.cool",
			"https://cats.cool",
			1,
			9,
			26,
		},
		{
			"https://lucky.me is a copy of dog.dev",
			"https://lucky.me",
			2,
			0,
			16,
		},
	}

	for _, tt := range tests {
		post, err := GenPost(tt.Input, []string{"en"})

		if err != nil {
			t.Errorf("input=%s failed: %s", tt.Input, err)
		}

		if len(post.Facets) != tt.ExpectedFacetCount {
			t.Errorf("got %d facets, expected %d",
				len(post.Facets), tt.ExpectedFacetCount)
		}

		facet := post.Facets[0]
		feature := facet.Features[0]

		if feature.RichtextFacet_Link.Uri != tt.ExpectedUrl {
			t.Errorf(
				"expected url=%s, got=%s",
				tt.ExpectedUrl,
				feature.RichtextFacet_Link.Uri,
			)
		}

		idx := facet.Index

		if idx.ByteStart != tt.ExpectedByteStart {
			t.Errorf("incorrect byte start: got=%d - expected %d",
				idx.ByteStart, tt.ExpectedByteStart)
		}

		if idx.ByteEnd != tt.ExpectedByteEnd {
			t.Errorf("incorrect byte end: got=%d - expected %d",
				idx.ByteEnd, tt.ExpectedByteEnd)
		}
	}
}
