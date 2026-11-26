---
type: result
status: preprint
area: structure
needs: none
measured-by: others
---
Threader keeps raw interactions and groups them into topic segments with a small BERT classifier. With 12 retrieved units, accuracy was 90.60 for segments, 87.00 with no segmentation, 85.60 for fixed chunks of size 500 and 85.80 for single messages. SeCom reports the same order on LoCoMo with BM25: topical segments 71.57, turns 65.58 and sessions 63.16, though its segment row also includes a compression-based denoiser. Both use a model to segment, so the result supports grouping turns by topic and not a model-free way of doing it.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Accuracy with 12 retrieved units (the figures match its LongMemEval results): segments 90.60, no segmentation 87.00, fixed chunks of size 500 85.60, single messages 85.80."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "SeCom on LoCoMo with BM25: topical segments 71.57, turns 65.58, sessions 63.16"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "the segment row also includes a compression-based denoiser, whose removal cost 9.46 points with a dense retriever"

## Relations
- supports:: [[Threads group turns by topic]]
- supports:: [[Atomicity has a floor]]
- source:: [[Threader]]
- source:: [[SeCom]]
