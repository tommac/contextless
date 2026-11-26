---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
Niksarli and Baheti store immutable identity nodes with version nodes that carry valid-time and transaction-time intervals. On 60 sampled LongMemEval questions recall at 10 was 46.7% overall and 80% on knowledge-update, but using the time-travel path lowered recall on temporal-reasoning questions from 50% to 37.5%, which the authors attribute to post-filter dilution. The sample is small and there is no baseline. The research note takes it as a reason to make time a soft or gated channel and not a hard post-filter.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "On 60 sampled LongMemEval questions R@10 was 46.7% overall and 80% on knowledge-update; using the time-travel path, recall on temporal-reasoning questions fell from 50% to 37.5%, which the authors attribute to post-filter dilution. No baseline."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "The bitemporal results found do not isolate the index from LLM extraction, and one shows time filtering lowering recall. Time should be a soft channel or a gated one, as in EdgeMem, not a hard post-filter."

## Relations
- supports:: [[Validity filters need an intent router for history questions]]
- part-of:: [[Bitemporal data model]]
- source:: [[LongMemEval]]
