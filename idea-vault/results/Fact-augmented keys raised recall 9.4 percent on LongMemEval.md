---
type: result
status: preprint
area: index
needs: none
measured-by: others
---
In the LongMemEval paper, indexing each round or session under extra keys made of facts extracted by a model, while returning the original text, raised Recall@5 on LongMemEval-M from 0.582 to 0.644 for rounds and from 0.706 to 0.732 for sessions. The authors report an average gain of 9.4 percent in recall and 5.4 percent in accuracy from fact keys. Round and session recall have different targets, so the table does not show sessions beating rounds.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Recall@5 / NDCG@5: rounds 0.582 / 0.481, rounds with fact-augmented keys 0.644 / 0.498, sessions 0.706 / 0.617, sessions with fact keys 0.732 / 0.620."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Fact keys gave an average 9.4% recall and 5.4% accuracy gain."

## Relations
- supports:: [[Deterministic keys on spans with verbatim output]]
- supports:: [[Atom titles as APIs]]
- supports:: [[Catalog cards are an index not a replacement]]
- supports:: [[Idea notes as verified keys over verbatim paragraphs]]
- source:: [[LongMemEval]]
