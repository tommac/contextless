---
type: idea
status: built
area: recall
needs: none
---
A long turn is cut by code at paragraph boundaries into parts of 24 KB, each with its own address of the form `tNNNNNN#pK`. Its catalog line gives its size and its parts, so a model can recall one part instead of the whole turn. The cuts are deterministic, so the same address always names the same text. The idea is ARC's chunked recall.

## Evidence
- `docs/ROADMAP.md`: "`tNNNNNN#pK`, deterministic paragraph cuts of 24 KB; catalog lines give size and parts; a search in a long turn returned the matching parts only"
- `docs/ARCHITECTURE.md`: "A long turn can be recalled one part at a time."
- `docs/ARCHITECTURE.md`: "cards for every turn not sent; a long turn's card gives its size and parts"

## Relations
- builds-on:: [[Recall by id or by search replaces memory]]
- builds-on:: [[Catalog cards are an index not a replacement]]
- analog-of:: [[Recall as paging]]
- source:: [[ARC Addressable Recall Compaction]]
