---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
On 100 stratified LongMemEval questions with an 8B reader, cosine-only retrieval gave F1 36.5 and BM25 plus cosine 44.2. By question type the gain was largest for knowledge updates, 36.3 to 63.0, and for single-session user questions, and slightly negative for multi-session and single-session assistant questions. Per-type samples are small, 11 to 27 questions, and the study is by the method's authors.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Cosine-only F1 36.5, BM25 + cosine 44.2 (+7.7)."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "knowledge-update +26.7 (36.3 → 63.0, n = 16), single-session-user +19.0 (n = 14), temporal +5.7 (n = 26), multi-session −2.2 (n = 27), single-session-assistant −3.0 (n = 11)"

## Relations
- supports:: [[Hybrid lexical and embedding retrieval beats either alone]]
- supports:: [[Fused ranking over exact full-text and value lanes]]
- source:: [[LongMemEval]]
