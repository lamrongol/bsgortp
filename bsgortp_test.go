package bsgortp

import (
	"testing"
)

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
	tests := []struct{
		Input string
		ExpectedUrl string
		ExpectedFacetCount int
	} {
			{
				"go visit https://cats.cool",
				"https://cats.cool",
				1,
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

		if post.Facets[0].Features[0].RichtextFacet_Link.Uri != tt.ExpectedUrl {
			t.Errorf(
				"expected url=%s, got=%s",
				tt.ExpectedUrl,
			 	post.Facets[0].Features[0].RichtextFacet_Link.Uri,
			)
		}
	}
}
