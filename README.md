# bsgortp - BlueSky Go Rich Text Parsing

Working on a ATProto project using Go and trying to create posts? It's kind of
a drag, huh? Say hello to **__bsgortp__**! Now you can efficiently turn that
little `string` into a full fledged `*bsky.FeedPost` with all the
`RichTextFacets` correctly populated.

## Installation

Getting started with `bsgortp` is very easy.

### Prerequisites

1) The Go toolchain
2) Working knowledge of ATProto and BlueSky (optional?)

Inside of your project just run:
```bash
go get github.com/jake-abed/bsgortp
```
Now you can import `github.com/jake-abed/bsgortp` into any file you need!

## Using bsgortp

Using bsgortp is comparably easy. Anywhere that you need a new `bsky.FeedPost`
struct, you could do something like this:
```go
post, err := bsgortp.GenPost(
    "Paging @jakeabed.dev... is this thing on? #test-post",
    []string{"en"},
)
```
If everything goes to plan, you'll get back the pointer to struct that looks
like this:
```go
bsky.FeedPost{
    LexiconTypeID: "app.bsky.feed.post",
	CreatedAt: "2025-03-11T04:20:00.156888Z",
	Facets []*RichtextFacet{ /* pointers to some RichtextFacets */ },
    Langs []string{"en"},
	Text "Paging @jakeabed.dev... is this thing on? #test-post",
}
```
The first Richtext Facet would look something like this:
```go
// Keep in mind most of these this are actually pointers to structs, but you
// get the idea
bsky.RichtextFacet{
    Features: []*RichtextFacet_Features_Elem{
        *RichtextFacet_Features_Elem{
            RichtextFacet_Mention: *RichtextFacet_Mention{
                LexiconTypeID: "app.bsky.richtext.facet#mention",
                Did: "some_did_i_will_not_type",
            },
        },
        Index: *RichtextFacet_ByteSlice{
            ByteStart: 8,
            ByteEnd: 20,
        },
    },
}
```
## To Do
- [ ] Extend test coverage
- [ ] Setup CI to run tests on merges to main
- [ ] Add embed support possibly
- [ ] Take in and modify/replace existing FeedPosts
- [ ] Improve Regex where needed
- [ ] Test & improve performance

## Contributing

Please feel free to open up an issue. Please do not create pull requests
without creatin an issue of some kind and discussing it first.

Eventually, CI will be setup to make sure all tests pass.
