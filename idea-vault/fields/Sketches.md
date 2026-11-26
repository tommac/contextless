---
type: field
status: external
area: index
needs: none
---
Sketches such as the count-min sketch, Bloom filters and MinHash answer questions about a large set in small space at the price of a bounded error. In agent memory the only use found is an unmeasured library feature, an existence filter that answers whether anything is known about a keyword before an expensive query. At single-conversation scale an exact inverted index already answers the same question at no extra cost, so sketches earn a place only across many conversations, where routing to the right conversation is the problem, or for collapsing near-duplicate tool output before packing. Loki's move from filters over text n-grams to filters over key-value pairs suggests that filters over typed values are the economical form.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "so sketches earn their place only across many conversations (open-memex shows recall@5 falling from 97.0% to 40.6% when all haystacks are pooled, which makes routing to the right conversation the problem)"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Loki's retreat from text n-gram filters to key-value filters suggests filters over typed values are the economical form."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No measurement of sketch-based existence tests in agent memory was found; Popoto's page gives configuration, not results."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "12x faster and 18x less disk on the peS2o corpus at the same deduplication quality"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "answers point, range and inner-product queries in sublinear space"

## Relations
- analog-of:: [[Catalog cards are an index not a replacement]]
- analog-of:: [[Digests label blocks of omitted turns]]
- alternative-to:: [[Zone maps and block-max indexes]]
