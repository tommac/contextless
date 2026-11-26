# One question about conflicting passages

On November 20, 2026, four isolated assistant sessions were given the question from **Idea workshop**. Each had to propose two distinct ideas, give one example and describe a cheap test. The following are their answers. No answers were graded blind. There was one generation per condition.

## Question only

> I’d choose the passage most likely to distinguish the two answers. For each candidate, use its speaker, timestamp, and location to predict what it would say if each answer were true. Prefer a passage whose predicted contents differ sharply and whose author could know firsthand. A second approach is to trace the disagreement to its source and open the original passage that one or both speakers were quoting. That can expose a missing condition.
>
> Suppose two messages list Tuesday and Thursday as the launch date. The remaining choices are a later team recap and the release manager’s decision immediately after a rescheduling vote. I’d open the decision. Approval makes Thursday current, rejection preserves Tuesday, and a deferral leaves the date unresolved. The recap could repeat either date without revealing the vote.
>
> To test that preference cheaply, hide the resolutions of 30 settled disagreements. Give each rule one passage and compare answer accuracy against always opening the newest passage.

## Question and labeled Canvas cards

> **1. Look for a decision that changes which answer applies.** If both passages concern the same person, event, and date, use the index to find a short reply from the decision-maker that confirms, cancels, or replaces one option. A later timestamp alone settles nothing.
>
> **2. Look for a condition that makes both answers true.** If the passages differ in audience, location, or phase, spend the read on the passage most likely to connect that difference to the question.
>
> Suppose a thread gives both Tuesday and Thursday for a workshop. The index lists an organizer’s reply titled “Confirmed date” and an undated agenda. I’d open the reply. “Thursday replaces Tuesday for everyone” settles the date. “Thursday is for instructors” leaves the attendee date unresolved.
>
> **Cheap test.** On 30 labeled conflicts, compare the decision-first rule with opening the newest related passage. Give each one equal-length read and score final answers. Drop the preference if it yields no more correct answers. The reported gain of 37 complete source sets measured retrieval.

## Same cards with relationship labels removed

> 1. **Read the passage where the two answers predict different outcomes.** Identify what would have to be true for each answer to hold. Use visible cues such as speaker, timestamp, and reply links to locate a passage likely to settle that fact.
> 2. **Expand the more ambiguous passage.** Open its immediate continuation when a missing qualification could reconcile the answers. “Friday” and “Monday” might refer to different offices.
>
> Suppose the organizer says “Keep Friday unless the venue confirms Monday.” A colleague later writes “I’ve penciled in Monday.” For the question “Which day is confirmed?”, I’d use the first idea and open the venue’s reply to the organizer. An acceptance selects Monday. A refusal leaves Friday in force. “We’ll check” leaves Friday as the standing plan and Monday pending.
>
> For a cheap test, hide the decisive replies in 30 labeled conflicts. Give each method the same passage index, one opening, and the same reader. Reject my preference for this batch if it resolves fewer cases correctly than adjacent expansion at the same abstention rate. The reported 37-case gain measured source coverage, so answer accuracy needs its own test.

## Ordinary vault search

> 1. **Look for a decision that distinguishes the answers.** Identify what would have to be true for each answer to hold. Use indexed titles, speakers, and thread pointers to find the passage most likely to settle that difference. Check its original wording before answering.
> 2. **Follow the conversation.** Find where the two accounts first diverge and open the next reply to that turn. A short “yes, for both offices” can settle a scope dispute even when it contains none of the search terms.
>
> Suppose one passage says a deadline is Friday and another says Monday. The Monday passage requests an extension. I’d use the first approach and open the project owner’s reply to that request. “Monday for the appendix only” supports Friday for the main report and Monday for the appendix. If the reply only acknowledges receipt, report the uncertainty.
>
> For a cheap test, label 30 deadline disputes and give each method one equal-length passage. Compare correct, source-supported answers against opening the most recent matching passage. Ties or losses would challenge my preference.

Search used the question text and the project's [flat vault search](../../docs/experiments/link_retrieval.py) over 524 notes. SQLite FTS5 BM25 ranked titles and opening prose with the experiment's Porter tokenizer and stopwords. The reader got the top three full notes with their paths, in rank order. The same three-card limit and 3,852-byte source-text ceiling applied to both selection methods. Search used 2,758 bytes. One note appeared in both selections.

1. [Excerpts separate retrieval text from evidence](<../../idea-vault/ideas/Excerpts separate retrieval text from evidence.md>)
2. [Growing excerpts reached 37 more complete source sets](<../../idea-vault/results/Growing excerpts reached 37 more complete source sets.md>)
3. [Write-time intelligence and read-time determinism](<../../idea-vault/ideas/Write-time intelligence and read-time determinism.md>)

The next experiment would use fresh design problems and repeat each condition with the same reader and output limit. Blind reviewers could score originality, feasibility and whether each proposed test would distinguish the idea from a simpler alternative. Record the time spent choosing notes as well.
