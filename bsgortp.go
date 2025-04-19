// The bsgortp package provides a simple and elegant way to quickly generate
// Rich Text Facets when working with
package bsgortp

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
)

// *TODO Improve and test regex for reliability and stability.
const LINK_EXP = `(http(s)?:\/\/.)?(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\` +
	`.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`

// Regex based on based on:
// https://atproto.com/specs/handle#handle-identifier-syntax as described on
// https://docs.bsky.app/docs/advanced-guides/post-richtext#rich-text-facets
const MENTION_EXP = `(@([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)` +
	`+[a-zA-Z]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)`

// Regex based on this stack overflow thread, but with modifications:
// https://stackoverflow.com/a/38383605
const TAG_EXP = `#(\w*[0-9a-zA-Z]+[\w\-]*[0-9a-zA-Z])`

const BSKY_BASE_URL = "https://bsky.social"

// FacetGenResult is a wrapper for the Rich Text Facets that facet generator
// functions output. If it was successful, the array will be filled and `Error`
// will be `nil`. Elsewise, the `Facets` field will be `nil` and `Error` will
// feature an `error` value.
type FacetGenResult struct {
	Facets []*bsky.RichtextFacet
	Error  error
}

// GenPost takes a string input as well as the applicable languages of the post
// and generates the pointer to a BlueSky FeedPost (as defined in the Indigo
// library).
//
// If generation fails for whatever reason, there are two possible paths:
// First, GenPost will return an error if the text is empty for some reason.
// Second, GenPost will return an error if any of the different types of facet
// generation fail.
func GenPost(text string, langs []string) (*bsky.FeedPost, error) {
	if text == "" {
		return nil, fmt.Errorf("cannot generate post with empty text")
	}

	now := time.Now().Format(time.RFC3339)

	facets, err := genFacets(text)
	if err != nil {
		return nil, err
	}

	post := &bsky.FeedPost{
		LexiconTypeID: "app.bsky.feed.post",
		CreatedAt:     now,
		Facets:        facets,
		Langs:         langs,
		Tags:          []string{},
		Text:          text,
	}

	return post, nil
}

// genFacets takes the text and delegates out responsibility to several other
// helper functions to parse.
func genFacets(text string) ([]*bsky.RichtextFacet, error) {
	facetChan := make(chan *FacetGenResult, 3)

	go genMentionFacets(text, facetChan)
	go genLinkFacets(text, facetChan)
	go genTagFacets(text, facetChan)

	facets := []*bsky.RichtextFacet{}

	for range 3 {
		facetGenResult := <-facetChan
		if facetGenResult.Error != nil {
			return nil, facetGenResult.Error
		}
		facets = append(facets, facetGenResult.Facets...)
	}

	return facets, nil
}

func genLinkFacets(text string, c chan<- *FacetGenResult) {
	r, err := regexp.Compile(LINK_EXP)
	if err != nil {
		err := fmt.Errorf(
			"could not compile link detection regex during facet generation",
		)
		c <- &FacetGenResult{Facets: nil, Error: err}
		return
	}

	urlMatches := r.FindAllString(text, -1)
	bytePositions := r.FindAllIndex([]byte(text), -1)

	if len(urlMatches) != len(bytePositions) {
		err := fmt.Errorf(
			"urlStrings=%v & bytePositions=%v not matched in facet generation\n",
			urlMatches,
			bytePositions,
		)
		c <- &FacetGenResult{Facets: nil, Error: err}
		return
	}

	facets := []*bsky.RichtextFacet{}

	for i := range urlMatches {
		facetLink := bsky.RichtextFacet_Link{
			LexiconTypeID: "abb.bsky.richtext.facet#link",
			Uri:           urlMatches[i],
		}

		facetElem := bsky.RichtextFacet_Features_Elem{
			RichtextFacet_Link: &facetLink,
		}

		facet := bsky.RichtextFacet{
			Features: []*bsky.RichtextFacet_Features_Elem{&facetElem},
			Index: &bsky.RichtextFacet_ByteSlice{
				ByteStart: int64(bytePositions[i][0]),
				ByteEnd:   int64(bytePositions[i][1]),
			},
		}

		facets = append(facets, &facet)
	}

	c <- &FacetGenResult{Facets: facets, Error: nil}
}

// genMentionFacets is the most tenuous of the three because it relies on a
// network request to resolve the provided Handle into a Did.
func genMentionFacets(text string, c chan<- *FacetGenResult) {
	r, err := regexp.Compile(MENTION_EXP)
	if err != nil {
		err := fmt.Errorf(
			"could not compile mention detection regex during facet generation",
		)
		c <- &FacetGenResult{Facets: nil, Error: err}
		return
	}

	mentionMatches := r.FindAllString(text, -1)
	bytePositions := r.FindAllIndex([]byte(text), -1)

	if len(mentionMatches) != len(bytePositions) {
		err := fmt.Errorf(
			"mentionStrings=%v & bytePositions=%v not matched in facet generation\n",
			mentionMatches,
			bytePositions,
		)
		c <- &FacetGenResult{Facets: nil, Error: err}
		return
	}

	facets := []*bsky.RichtextFacet{}
	client := &xrpc.Client{
		Host: BSKY_BASE_URL,
	}

	for i := range mentionMatches {
		// Crucial detail here is that each of the matches will return the '@'
		// character as their first byte/rune, so you need to trim it off when
		// resolving the handle otherwise you will get false negatives.
		handleOutput, err := atproto.IdentityResolveHandle(
			context.Background(), client, mentionMatches[i][1:])
		if err != nil {
			// Wrap the error and note the handle that could not resolve
			wrappedErr := fmt.Errorf(
				"could not resolve handle=%s, error : %w",
				mentionMatches[i],
				err,
			)
			c <- &FacetGenResult{Facets: nil, Error: wrappedErr}
			return
		}

		facetMention := bsky.RichtextFacet_Mention{
			LexiconTypeID: "abb.bsky.richtext.facet#mention",
			Did:           handleOutput.Did,
		}

		facetElem := bsky.RichtextFacet_Features_Elem{
			RichtextFacet_Mention: &facetMention,
		}

		facet := bsky.RichtextFacet{
			Features: []*bsky.RichtextFacet_Features_Elem{&facetElem},
			Index: &bsky.RichtextFacet_ByteSlice{
				ByteStart: int64(bytePositions[i][0]),
				ByteEnd:   int64(bytePositions[i][1]),
			},
		}

		facets = append(facets, &facet)
	}

	c <- &FacetGenResult{Facets: facets, Error: nil}
}

func genTagFacets(text string, c chan<- *FacetGenResult) {
	r, err := regexp.Compile(TAG_EXP)
	if err != nil {
		err := fmt.Errorf(
			"could not compile tag detection regex during facet generation",
		)
		c <- &FacetGenResult{Facets: nil, Error: err}
		return
	}

	tagMatches := r.FindAllString(text, -1)
	bytePositions := r.FindAllIndex([]byte(text), -1)

	if len(tagMatches) != len(bytePositions) {
		err := fmt.Errorf(
			"tagStrings=%v & bytePositions=%v not matched in facet generation\n",
			tagMatches,
			bytePositions,
		)
		c <- &FacetGenResult{Facets: nil, Error: err}
		return
	}

	facets := []*bsky.RichtextFacet{}

	for i := range tagMatches {
		facetTag := bsky.RichtextFacet_Tag{
			LexiconTypeID: "abb.bsky.richtext.facet#tag",
			Tag:           tagMatches[i][1:],
		}

		facetElem := bsky.RichtextFacet_Features_Elem{
			RichtextFacet_Tag: &facetTag,
		}

		facet := bsky.RichtextFacet{
			Features: []*bsky.RichtextFacet_Features_Elem{&facetElem},
			Index: &bsky.RichtextFacet_ByteSlice{
				ByteStart: int64(bytePositions[i][0]),
				ByteEnd:   int64(bytePositions[i][1]),
			},
		}

		facets = append(facets, &facet)
	}

	c <- &FacetGenResult{Facets: facets, Error: nil}
}
