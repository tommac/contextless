# Context compiler: atoms, levels of detail and budgeted packets

Design proposal, 2026-11-15. Phase 1 is built (section 2b says what it
measured, which is less than section 2 promised); the rest is not.

## 0. In short

- **Break every turn into atoms**: paragraph-sized spans of the verbatim
  turn (about 50 to 600 tokens), each with a stable id such as
  `t000123#a4`. An atom is an address into the immutable turn, never a
  rewrite or a summary. This is the "every paragraph is an Obsidian note"
  idea, kept lossless.
- **Measured on three real conversations** (90 probes): with atoms, the
  answer enters the packet after a median of 267 to 339 tokens, against
  1,718 to 2,077 for whole turns, about six times fewer. Within 2,000
  tokens, atoms hold the answer for 26 of 30 probes on each conversation;
  whole turns for 14 to 17.
- **Built into packets, the gain is smaller** (section 2b). A packet
  fills a fixed budget, and harder questions need more of it: at the 8k
  default, atoms found 136 of 150 answers against 130 for whole turns, in
  packets the same size to 41 percent smaller. At a third of the tokens
  they were three answers behind.
- **Build each packet like a database query planner**: candidate lanes,
  expansion, budgeted selection, a layout that keeps the prompt cache warm,
  then a cost and time estimate per provider. `cx explain` prints the plan.
- **Show history at levels of detail**, like map tiles: the question and
  the latest exchanges in full, the hits as atoms, everything else as cards
  and digests, every level addressable for recall.
- **Learn from signals that are already free**: the model's recall
  requests (misses), its citations (hits) and the probes (offline) tune the
  weights per workspace. No model runs at selection time.

## 1. Goals and constraints

Three things to optimize, in this order:

1. **Quality**: the evidence the answer needs is in the packet, readable.
2. **Cost**: fewer billed input tokens, more of them read from the prompt
   cache, fewer output tokens.
3. **Time**: shorter time to first token, fewer recall round trips.

Kept from the current design: turns are verbatim and never rewritten;
selection is done by code; everything left out is named and recallable;
model calls happen only at index time, are optional, and their output is
an index, never a stand-in for the text; every change is measured before
it is kept.

Not proposed: prompt compression (LLMLingua and similar), summaries in
place of text, embeddings before you approve a local model download.

## 2. Measured first: does finer granularity help?

**Setup.** The three workspaces and 90 probes from the real-transcript run
(see ARCHITECTURE.md 3c). Each probe is a question with its source turn
(and other turns that restate the answer). For each probe, units are
ranked by SQLite FTS5 bm25 with the same query rule as `cx` (lowercase
words of three or more letters, stopwords removed, OR). Units are added to
the packet in rank order, and we record how many tokens are spent before a
unit that holds the answer is in. One lane only (full text), no fusion, no
model. Scripts: [docs/experiments/atoms_exp.py](experiments/atoms_exp.py)
and [size_sweep.py](experiments/size_sweep.py) (Python with SQLite FTS5;
run with the scratch directory that holds the workspaces and probe files).
They would become a `cx probe -granularity` mode.

Atoms here: blocks split at blank lines, code fences kept whole, blocks
under about 50 tokens joined to the next, blocks over about 600 tokens cut
at a line or sentence end. Each atom costs its tokens plus a 25-token
header.

| Conversation | Units | Answer within 2k | within 4k | within 8k | Median tokens to answer |
|---|---|---|---|---|---|
| BookWriter, 140 turns | whole turns | 17 of 30 | 20 | 24 | 1,718 |
| | atoms | 26 of 30 | 28 | 29 | 267 |
| BookWriter + 8 sessions, 299 turns | whole turns | 16 of 30 | 17 | 24 | 1,760 |
| | atoms | 26 of 30 | 27 | 28 | 270 |
| Research session, 201 turns | whole turns | 14 of 30 | 20 | 24 | 2,077 |
| | atoms | 26 of 30 | 29 | 29 | 339 |

Whole turns use the plain tokenizer (today's index); atoms use the porter
stemmer. Each is the better of the two tokenizers for its unit.

What else the runs showed:

- **Any atom size beats whole turns.** Sizes from about 20 to 2,400
  tokens per atom all did; smaller atoms needed fewer tokens, and the gain
  levels off below about 600.
- **Stemming had small, mixed effects**: it lowered the median cost of
  atoms by 6 to 17 percent and raised that of whole turns by 2 to 25
  percent (a long turn matches more stray word forms). It belongs with
  atoms, if anywhere.
- **Adding each atom's neighbors** (the paragraph before and after)
  doubled the median cost. Recovery was unchanged within 8k and 16k, but
  fell by 1 to 3 of 30 within 2k and by 1 to 4 of 30 within 4k, because
  the same budget holds about half as many hits. It
  is still 2.5 to 3.5 times cheaper than whole turns, and it is the price
  of readability, so end-to-end tests decide it (section 9).
- **Indexing each assistant atom with the question it answers** did not
  help overall (a little better on one conversation, worse on another):
  these probe questions already share words with their answers.

How this maps onto today's packets (dry runs with one probe question;
planned totals also hold about 15k of CLI system prompt and a 25k output
reserve):

| Conversation | Brief | Ranked turns (6) | Catalog | Recent turns (8) |
|---|---|---|---|---|
| BookWriter, 140 turns | 22,701 | 11,958 | 5,734 | 4,689 |
| BookWriter + 8 sessions | 26,984 | 8,825 | 12,764 | 5,380 |
| Research session | 52,077 | 37,600 | 8,935 | 5,388 |

A fixed 4,000-token atom budget would replace sections costing 8,800 to
37,600 tokens, and in the experiment it held the answer for 27 to 29 of
30 probes. The catalog is the next largest section; levels of detail and
the prompt cache (sections 3.3 and 5) address it.

**Caveats.** One lane, 90 probes, questions written by a model that had
read the source, so they share rare words with the answer. "The answer is
in the packet" is not "the model answers correctly": a paragraph out of
context can lose what "it" refers to. That readability risk is the main
open question, and only an end-to-end test with a model settles it.

## 2b. Phase 1 as built: what the probes say

Built: the atom splitter (`vault.Atoms`), a stemmed atom index that also
searches each atom with its turn's card text, `tNNNNNN#aK` and `#aK-M`
recall ids, ranked atoms sent as excerpts under a token budget
(`relevant_tokens`, default 8,000) in place of six ranked turns, recall
search that returns paragraphs, the whole history when it is short
(`whole_history_tokens`, default 60,000, and only if it fits the first
profile), and `cx probe -granularity atoms|turns -budget N`.

Measured with `cx probe` on the real conversations, whole turns against
atoms at three budgets. The three workspaces were imported without
running the indexer, so they have no cards, threads, facts or decisions;
nothing below is measured on an indexed workspace yet. "Old" probes are
the 90 from the first real run; "fresh" are 60 new ones a sub-agent wrote
for this test without reading any code, phrased as vaguer memories of the
facts. The two BookWriter old rows ask the same 30 questions, on the
140-turn session and on the 299 turns that begin with it, so the 150
probe runs are 120 distinct questions. Source turn (or the excerpt
holding the answer) in the packet, out of 30:

| Conversation, probes | 6 whole turns | Atoms, 4k budget | Atoms, 8k | Atoms, 16k |
|---|---|---|---|---|
| BookWriter 140 turns, old | 27 | 28 | 29 | 29 |
| BookWriter 299 turns, old | 26 | 27 | 28 | 29 |
| Research 201 turns, old | 29 | 27 | 27 | 28 |
| BookWriter 299 turns, fresh | 25 | 22 | 26 | 26 |
| Research 201 turns, fresh | 23 | 23 | 26 | 26 |
| All 150 runs | 130 | 127 | 136 | 138 |
| Distinct questions, 120 | 103 | 99 | 107 | 109 |
| Fresh 60 only | 48 | 45 | 52 | 52 |

Counting any turn that states the answer (the probes' `also` lists), the
five sets read 27, 26, 29, 26, 23 for whole turns and 29, 28, 29, 26, 26
for atoms at 8k. That column undercounts both: an `also` turn is credited
only when it holds the source's exact wording.

Median tokens of the ranked section: whole turns 9,700 to 25,700 by
conversation; atoms about 3,900, 7,700 and 15,500 at the three budgets
(the budget counts headers, so the section stays under it). Median brief
at the 8k budget against whole turns: 18,800 against 24,100, 26,200
against 28,000, and 22,700 against 40,300 on the research session, whose
turns are long.

These numbers are from the build after an independent review of Phase 1
(58 findings, all checked and the defects fixed). The review changed the
splitting rules (a heading opens its section's atom), the budget
accounting and the exact lane; recovery totals did not move (130, 127,
136, 138 before and after), single sets moved by one probe.

What this says, plainly:

- **At the 8k default, atoms find a few more answers (107 against 103 of
  the 120 distinct questions; 52 against 48 on the fresh probes) in
  packets 6 to 44 percent smaller.** That is within noise: per
  conversation the differences are one to three probes of 30, and a
  paired sign test on the distinct questions (atoms won 11, lost 7 on
  the build before the review) gives p about 0.5. The honest reading is
  "no worse, for less", not "better".
- **The sixfold saving of section 2 did not carry over.** That figure was
  the cost of reaching the answer in rank order. A packet fills a fixed
  budget whatever the question, and the vaguer fresh questions need more
  of it. At about a third of the ranked tokens (the 4k budget) atoms are
  three probes behind whole turns overall and three behind on the fresh
  ones, also within noise. Phase 1's own bar, "at least match turn probes
  at a third of the tokens, on fresh probes", was therefore missed as
  written; atoms are kept because at 8k they do no worse for less.
- **Recall search needs more paragraphs than turns.** The first build
  returned one paragraph where whole-turn mode returned one turn, and
  its probe column counted any paragraph of the source turn as a hit.
  Counted by whether the returned text holds the answer, that was 108 of
  150 against 126 for whole turns. A search now returns four paragraphs
  per turn it would have returned: 127 of 150 (28, 27, 27, 22, 23 by
  set, against 27, 25, 29, 23, 22). The probe searches at depth five;
  the engine puts back three hits per search, so real recall is lower
  than both figures.
- **The value lane does not transfer to atoms.** Fused as a full lane it
  cost up to 5 of 30 answers on the fresh probes (26 down to 21 and 22):
  among thousands of short atoms, a number near a topic word is common,
  and every atom the lane brings in takes budget from the full-text
  lane's better ones. Keeping only its two best atoms never cost an
  answer on any set. That setting was chosen after seeing these probes,
  the fresh ones included, so they are no longer held out for it (and it
  was measured on the build before the review).
- **Cards matter to atoms.** On the synthetic conversation, atoms lost a
  planted decision until each atom was searched with its turn's card
  text; the question shared its words with the card, not with the
  paragraph. That is the "titles as APIs" idea of section 3.1 at turn
  level, and the first evidence for doing it per atom. A card match says
  which turn, not which paragraph, so a turn found only through its card
  goes in whole when it fits the budget and is left to its catalog line
  when it does not.
- **Where threads exist, the budget is not the whole story.** A ranked
  atom inside a thread turn that is already in the packet costs nothing,
  and the turn is kept whole as a ranked turn. On the indexed synthetic
  conversation most ranked text is such thread turns (a median of 45,600
  tokens reported as ranked at the 8k budget), and the brief is the same
  size as with whole turns. The size result above is for conversations
  with no threads.
- **The whole history wins when it fits.** With the limit raised to
  100,000 tokens, the 140-turn conversation goes out whole (an 83,100
  token brief) and all 30 sources are in it, against 29 at 18,800 tokens
  with atoms. The default limit of 60,000 is a cost choice, not a
  measured optimum. Plan units and their assembler are excluded, so a
  plan does not send the transcript once per unit.
- Still unmeasured: whether a model reads an excerpt as well as a whole
  turn. Excerpts carry the turn, the paragraph numbers and the message
  they answer in their header, and the paragraphs around one come back by
  id, but only an end-to-end run settles it.

Known gaps the review found that are not fixed yet:

- An excerpt cut from a long code block leaves a code fence open in the
  brief. Its header says so, but a Markdown renderer would still read the
  rest of the brief as code.
- A reviewer run cannot recall. It now gets the records its writer
  recalled, and is told to judge with what it has.

Reproduce: `cx probe <ws> <probes.jsonl> -whole -1 -granularity turns`,
then `-granularity atoms -budget 4000` (8000, 16000); `-value-top N` and
`-lanes` vary the lanes. The probe files hold private conversation text
and are not in the repository.

## 2c. The offline floor

`cx floor` measures every baseline on the same questions with no model:
is the answer's source text in the packet, how much ranked text comes
before it, what the packets cost, and how long they take to build. On the
120 distinct real questions, on workspaces with no cards or threads:

| Arm | Answer in packet | Against current at 8k: won, lost, exact p | Median brief (tokens) | Input per replayed message |
|---|---|---|---|---|
| Whole history | 120 | 13, 0, under 0.001 | 297,525 | 219,541 |
| Recent turns only | 0 | 0, 107, under 0.001 | 18,616 | 15,345 |
| Six whole turns | 103 | 7, 11, 0.48 | 31,002 | 29,382 |
| Full text only, or all lanes, at 2k | 91 | 0, 16, under 0.001 | 20,200 | 17,000 |
| The same at 4k | 99 | 0, 8, 0.008 | 22,100 | 18,800 |
| The same at 8k | 107 | reference | 25,913 | 22,260 |
| The same at 16k | 109 | 2, 0, 0.50 | 33,010 | 29,060 |
| The same at 32k | 113 | 6, 0, 0.03 | 48,121 | 41,070 |

- **The exact and value lanes change no packet result here.** Full text
  alone found the same number at every budget. They may matter for other
  questions, but these 120 give them no credit; the median ranked text
  before the answer is 400 tokens with full text alone and 432 with all
  lanes.
- **More budget buys little after 8k:** four times the ranked text adds six
  answers. Whole history finds all 120 at about ten times the input.
- These probes place the answer in the packet. They do not show a model
  uses it, and none of this has been run with a model.
- 32 `also` entries name a turn without the probe's exact wording and cannot
  count. Probes now take `answers`, other exact wordings an `also` turn may
  hold; the old files have not been rewritten.

Real calls now build the same packet as `cx packet` and `cx probe`: a
message enters the index after it is answered, so its words no longer shift
the statistics of its own ranking (28 of 150 ranked sections differed).

**Message timing.** In the 10 source sessions, 254 gaps between user messages
have a median of 16 minutes; 37 percent are within 5 minutes, 59 percent
within 30, 68 percent within an hour and 94 percent within 12 hours. `cx gaps`
turns these into the price of a stable prompt prefix per message, as a multiple
of the normal input price (reads at 0.1): 0.82 for a 5-minute cache with writes
at 1.25, 0.57 for 30 minutes at 1.25, 0.71 for 1 hour at 2, and 0.24 for 12
hours at 2.5. Keep-alive requests on the 5-minute cache do not pay at this pace.
The figures cover the stable prefix only, one call per message. Sessions
imported from now on keep their message times.

**New held-out probes** (owner's task): at least 100 questions written from
memory without viewing the transcript, then matched to a source, one JSON
object per line with `id`, `type`, `question`, `source`, `answer` and
optionally `also` and `answers`. Keep them outside the repository, and do not
tune against them.

Reproduce: `cx floor <ws>=<probes.jsonl>...`, and `cx gaps <session.jsonl>...`.

## 2d. Reaching passages spread over a conversation

Questions such as "how did this decision change?" need several passages far
apart. `cx zoom` measures, with no model, how many needed passages each way of
choosing text puts within a token total. A passage counts when its exact words
are in what was read. **This says what a packet holds, not whether a model
answers well.**

Idea notes are one-line statements written by a model that read the
conversation without seeing any test question, each with quotes that code
checks against the cited turn and pins to paragraphs. Notes are used as extra
search keys: a matching note passes its rank to its paragraphs, and with topics
the best notes bring in the other notes filed under the same subject. What is
read is always the verbatim paragraph. All 1,117 quotes in 955 notes were found.

**Held-out result.** Settings were chosen on a development set only, and the
pass rule was written first: significantly ahead of plain search (exact sign
test, p below 0.05) at 3 or more of the 4 budgets on whole-conversation
questions, with no significant loss on one-source questions. The run used 150
whole-conversation questions (90 on a conversation never used for tuning, 60
newly written on the development conversation) and 60 one-source questions.
Share of needed passages read:

| Way of choosing | 4,000 tokens | 8,000 | 16,000 | 32,000 | Budgets significantly ahead | Rule |
|---|---|---|---|---|---|---|
| Plain search | 19.6 | 27.6 | 39.3 | 54.5 | | baseline |
| Search + notes | 22.4 | 33.7 | 46.3 | 61.0 | 3 | passes |
| Search + notes + topics | 24.0 | 33.7 | 48.0 | 61.8 | 3 | passes |
| Same, with neighbouring paragraphs | 17.4 | 30.5 | 46.3 | 62.7 | 2 | fails |
| Search + links through rare words | 19.6 | 28.8 | 40.9 | 56.9 | 2 | fails |
| Search + feedback (borrowed words) | 20.4 | 29.0 | 39.6 | 54.9 | 0 | fails |
| Search + coverage picking | 18.3 | 26.0 | 38.4 | 54.0 | 0 | fails |
| Search + neighbouring paragraphs | 14.0 | 25.2 | 38.0 | 54.6 | 0 | fails |
| Note lines shown, rest by search | 14.7 | 24.3 | 34.5 | 51.7 | 0 | fails, with one-source losses |
| Code-built zoom ladder, keyword reader | 2.6 | 6.0 | 10.6 | 18.9 | 0 | fails |

Limits on this result:

- The two passing arms are 2 of 8 tested with no correction for multiple tests.
- They do not pass everywhere. On the never-tuned conversation alone, search +
  notes passes (3 budgets) and notes + topics does not (2). On the 60 new
  questions from the development conversation neither is significant at any
  budget, though both lead at all four (for example 23 won, 14 lost at 8,000).
- At 4,000 tokens the notes arms lose ground on one-source questions: 76.7
  percent against 85.0 for search (not significant, 4 won and 9 lost for search
  + notes).
- The ladder result is about this keyword reader, which opens only lines
  sharing a word with the question. No model has read a ladder here. Its arms
  also spend less than the total (a mean of 2,448 of 4,000 tokens).
- The variants' settings were the best of a small sweep on development data.

**Answer pilots: not evidence yet.** Two small runs put packets in front of a
model, with blind judges and a whole-conversation arm. Both used development
questions only: 4 under an earlier packet format, then 3 under the current one.
On those 3 questions, the share of needed points stated was 0.47 for search and
0.53 for notes at 8,000 tokens, 0.73 for every note line alone, 0.80 for the
whole conversation (2 questions), 0.93 when the model read every note line and
then asked for paragraphs by id, and 1.00 when it did the same from code-built
lines per turn. Three questions cannot rank these arms. They do suggest what
the full test should look at first: **an overview plus a model's own request for
paragraphs**, including an overview that needs no model to write. In that pilot
every passage held in a packet of limited size was stated (with the whole
conversation, 80 percent were), and passages neither held nor pointed at almost
never were, so what reaches the model seems to decide the answer. The full test
(150 questions) has not been run.

Reproduce: `cx zoom <ws>=<questions.jsonl>... -notes DIR -out DIR`; `-dump DIR`
writes the packets for an answer test and `cx pull` serves paragraph requests.

## 2e. The answer test of 2026-11-25, and the map in live packets

The test section 2d ends on was run on 2026-11-25, under a rule written on
2026-11-18 before any of its answers existed. **This is the first result here
about answers and not about what a packet holds.**

Thirty held-out questions that need several passages far apart (every fifth
one: 18 on the conversation never used for tuning, 12 written later on the
development conversation). One answering model (Claude Opus 5.5, medium
effort), one call per packet, two for the arms that ask for paragraphs. Two
judges on different models marked each needed passage yes, partial or no for
each answer, under arbitrary labels and in opposite orders. They agreed
exactly on 0.936 of 965 verdicts (Cohen's kappa 0.889). 255 agents ran with no
error, and the audit dropped no answer. The score is the share of needed
passages both judges' average marks as stated.

| Arm | What the model read | Score | Tokens read |
|---|---|---|---|
| none | the question only (control) | 0.000 | 130 |
| search | paragraphs by plain search | 0.314 | 7,961 |
| searchbig | plain search, filled to the size of the note packets | 0.546 | 34,776 |
| catpull | a code-built line per turn, then the paragraphs it asked for | 0.545 | 23,111 |
| pull | every checked note line, then the paragraphs it asked for | 0.726 | 31,001 |
| full | the whole conversation (15 of the 30 questions) | 0.839 | 247,856 |
| oracle | exactly the needed paragraphs (control) | 0.997 | 1,209 |

**The rule's verdict: the lead it tested does not hold.** The rule asked
whether an overview that needs no model to write (catpull) beats search of the
same size. It does not: 12 questions won, 12 lost, 6 tied, a mean difference
of -0.001 (95 percent interval -0.125 to +0.122). Both controls passed (none
at most 0.15, oracle at least 0.85), so the test is trusted.

**What the same run shows about model-written notes.** These comparisons were
listed in the plan as secondary, with Holm's adjustment across them, and the
rule of 2026-11-25 calls them description, so they are a strong lead and not
a confirmed result:

- pull against searchbig: +0.180 (interval +0.080 to +0.281), 18 won, 5 lost,
  7 tied, exact sign p 0.011, adjusted 0.032. On each conversation alone the
  lead is the same size and not significant (adjusted 0.18 on the never-tuned
  one, with 18 questions).
- pull against catpull: +0.181 (interval +0.085 to +0.273), 21 won, 7 lost,
  adjusted p 0.032. The note lines matter; a line per turn made by code is not
  a substitute.
- full against pull, on the 15 questions that have a whole-conversation
  answer: +0.070 (interval -0.060 to +0.197), 6 won, 2 lost, 7 tied, p 0.29.
  No difference was detected at one eighth of the tokens, but 15 questions
  cannot exclude a gap of up to 0.2.
- oracle against full: +0.161 (interval +0.059 to +0.286), 8 won, none lost.
  A model with the whole conversation in front of it left out passages that it
  stated every time when they were all it was given. Sending everything is not
  the ceiling.

**What reaches the model decides the answer.** Over all arms, a needed passage
that the packet held word for word was marked stated in 0.92 to 1.00 of judge
verdicts (0.83 with the whole conversation). One that was neither held nor
pointed at was stated in 0.09 to 0.28. One that a shown note line pointed at
but the model did not read was stated in about a third (0.33 and 0.38, on 36
and 8 verdicts). A note line is a pointer, not a substitute for the paragraph.

**Where pull fell short**, counted afterwards with no model, so description
only. It held 92 of the 148 needed passages. Of the 56 it missed, 22 had a
note line pointing at them that the model did not ask for, and 34 had no note
at all. The model asked for little: a mean of 21 paragraphs and 3,963 of the
8,000 tokens allowed, and never for more than fitted (catpull asked for more
than fitted in 70 percent of its requests). The two overviews found different
passages: the paragraphs of the pull and catpull requests together held 113,
and pull's with an ordinary 8,000-token search packet 106.

Limits: 30 questions on two conversations of one user, one answering model,
and notes written in a separate pass by a model whose cost is in none of the
token columns (it is paid once, when turns are indexed).

### The map of the conversation, as built (2026-11-25)

The pull arm is now a path of the engine, called the atlas:

- **Notes are records.** A note (`notes/<id>.md` in the vault) holds a title,
  a kind, topics, and evidence: for each quote, the turn, the byte span, the
  paragraphs that hold it and the hash of the body it was found in. It is
  stored only if code finds at least one of its quotes in the turn it cites,
  and its id is a hash of its title and spans, so the same note written twice
  is one record. `cx notes <name> import` stores notes written elsewhere after
  that check; with `atlas: on` the indexer writes them for each batch it
  indexes, shown the topic names already in use.
- **The packet shows them as a map**, one line per note in the format the test
  used (`- [t000012#a3] (decision) title`), after the decisions in force and
  before everything chosen for the message. Notes are only ever added, so while
  the map fits its budget (`atlas_map_tokens`, 40,000) it is the same for every
  message and grows at its end. A note is shown only through evidence that lies
  before the message being answered, on turns whose text is still what was
  checked. A packet that holds every turn has no map.
- **The model asks, the engine serves.** The section tells the reader that a
  note points and is not the text, and to ask for the paragraphs an answer
  rests on with `needs_recall`. While a map is shown a request is bounded by
  tokens (`atlas_pull_tokens`, 8,000) instead of by count: ids are served in
  the order asked, one that does not fit is passed over for a smaller one, and
  a note in the next packet names what was left out. A second request gets the
  same room again.
- **Past the budget nothing disappears without a line.** The lines ranked for
  the message are shown, and the rest are counted under their topic with their
  turn range; the recall op `notes` lists a topic's lines, or the best matches
  of some words.

`cx notes <name>` reports how much of a conversation its notes point at and
what the map costs; `cx packet <name> -pull id,id -ask "words;words" <message>`
replays a request (ids, searches or both) without a model. While the map is
shown, what a search puts back is paid for from the same tokens as the ids.

**Checked without a model** (the 30 questions above, so this is a check that
the live path does what the tested arm did, not a new result):

- All 955 notes of the test were imported with every one of their 1,117 quotes
  found. They point at 28 and 27 percent of the two conversations' paragraphs,
  in 77 and 64 percent of their turns, and their lines cost 10 and 12 percent
  of the conversations' tokens.
- Replaying each of the 30 logged requests through the live packet builder,
  the packet held 108 of the 148 needed passages word for word, against 92 in
  the tested arm, more on 12 questions and fewer on none. The live packet also
  carries the recent turns and 8,000 tokens of ranked paragraphs, which is
  where the extra passages come from. Before any request it held 39.
- The price: a mean brief of 51,111 tokens before the request and 54,958
  after, against 31,001 for the tested arm.
- `cx cache` on the last 40 messages of each conversation: with the map the
  median brief is 47,960 and 41,185 tokens with 61 and 46 percent of it in a
  prefix a provider's cache could serve; without it 19,108 and 22,311 tokens
  with 12 and 4 percent. Billed-equivalent input per message, if every call
  hits the cache: 22,974 against 18,550, and 24,084 against 21,615. If none
  does: 49,723 against 20,805, and 41,364 against 22,603. Whether the CLIs
  reuse a prefix is still unproven (section 5), so the map costs between a
  tenth more and more than twice as much.

**Not measured, each a reason the default is cautious** (`atlas: auto` shows a
map only where notes exist, and the indexer writes none unless asked):

- No model has answered from a live packet with a map. In the test the request
  was a required step; in a live packet the model decides whether to ask, and
  may answer from the note lines, which the test says is worth a third of a
  passage.
- No note has been written by the indexer under the new instructions. The
  tested notes came from separate writers that read long stretches at once.
- The map past its budget (ranked lines, topics named, the `notes` op) has not
  been in front of a model.
- The 34 passages no note covered are the larger share of what pull missed.
  A map with a line for turns that have no note is the obvious next arm.

The first of these has since been tested (section 2f): a model answered from
live packets with the map, with the request offered as a step.

## 2f. The confirming test of 2026-11-25: the map in live packets

Section 2e ends with a lead from a secondary comparison and an engine feature
no model had answered from. The same day a second test put both to a rule
written before any of its packets or answers existed.

**Two tests, on the 120 held-out questions the first run did not use** (72 of
the held-out conversation, 48 of the other), Holm-adjusted across the two,
two-sided exact sign test on per-question strict score:

- Q1, the earlier finding repeated as it was run: note lines plus one request
  for paragraphs (pull) against search of the same size (searchbig). Same
  binary, packet code and prompts as the first run; with that binary the first
  run's 330 packets regenerate byte for byte.
- Q2, the engine: the packet `cx` builds for the question as the next message
  of the conversation, with the notes imported and default settings, against
  the same engine with `atlas: off`. Each arm may make one recall request
  through the engine's own recall path: ids, up to 5 searches, or both. One
  prompt for both arms.

Controls (question alone, exact paragraphs) and the whole conversation ran on
30 of the 120. Answers by Opus 5.5 at medium effort; judges Opus 5.5 and
Sonnet 5.5, blind, opposite orders: the models of the first run. 810 calls, no
errors.

| What the model read | n | Needed passages stated | Tokens |
|---|---|---|---|
| Question alone | 30 | 0.000 | 130 |
| Search filled to the notes' size (searchbig) | 120 | 0.672 | 34,768 |
| Engine packet, map off, one request (livenomap) | 119 | 0.706 | 28,984 |
| Note lines, then requested paragraphs (pull) | 120 | 0.762 | 31,415 |
| Whole conversation | 30 | 0.810 | 247,854 |
| Engine packet with the map, one request (livemap) | 119 | 0.882 | 58,517 |
| Exactly the needed paragraphs | 30 | 0.982 | 1,214 |

**Verdicts under the fixed rule.**

- Q1 is CONFIRMED: pull is ahead of searchbig by 0.090 (95% interval +0.037 to
  +0.141; 57 won, 30 lost, 33 tied; adjusted p 0.005). The first run's +0.180
  on 30 questions was about twice the size this run finds.
- Q2: THE MAP HELPS THE ENGINE. The live packet with the map is ahead of the
  same engine without it by 0.175 (95% interval +0.137 to +0.215; 66 won, 4
  lost, 49 tied; adjusted p below 0.0001).
- Trust conditions held: controls 0.000 and 0.982, judge kappa 0.840 over
  2,781 verdicts, and the audit dropped 2 of 570 answers (one question, both
  live arms, for writing a one-word search without quotes; with them counted
  the numbers are the same to the second decimal).
- No qualification: contradictions flagged by both judges were 0 for livemap
  and 2 for livenomap, 3 for pull and 4 for searchbig.

**Described, not tested.**

- The live packet with the map against the whole conversation, on the 29
  questions both have: 0.035 ahead (95% interval -0.039 to +0.118), at under a
  quarter of the tokens. No difference detected, in either direction.
- The live packet with the map against pull: +0.118 (56 won, 10 lost), and
  against searchbig: +0.205 (77 won, 11 lost).
- The engine without the map against searchbig: +0.030 (interval -0.019 to
  +0.080). One request with searches did not lift it past plain search.
- Each conversation apart: the map's gain in the engine is +0.175 and +0.176.
  Pull against searchbig is +0.119 on the held-out conversation and +0.047
  (interval -0.045 to +0.134) on the other, so Q1's confirmation rests mostly
  on one of the two.
- The whole conversation scored 0.719 on the longer conversation (295,000
  tokens) and 0.946 on the shorter (176,000). The exact packet beat it by
  0.172 (16 won, 1 lost): the first run's finding, repeated.

**Why the map helps: what reached the model.** Before any request both live
packets held 0.327 of the 588 needed passages word for word. After the model's
one request, the packet with the map held 0.840 and the one without 0.614.
With the map the model asked for a mean of 10.3 ids and 3.2 searches and got
7,444 tokens back; without it, 4.0 ids and 4.3 searches and 4,953 tokens. A
held passage was stated in 0.95 to 0.97 of verdicts in every arm but the whole
conversation (0.80). So the gain is the model choosing well from the map, not
answering from note lines.

**The test's shortcut, checked.** The test's model kept its first packet and
read what the request put back, where a real second call gets one rebuilt
packet. Replaying all 240 logged requests through the packet builder with no
model: every needed passage the test's model had is in the real second packet
(0 lost; 1 and 4 more held), no second packet was too big, and each request
replayed identically.

**Limits.**

- The live packet with the map is twice the size of the one without (58,517
  against 28,984 tokens). No arm fills the map's tokens with search results at
  that size. What the run does show: search at 34,768 tokens scored below the
  map-off engine at 28,984, and the whole conversation at 247,854 scored below
  the map, so more raw text alone is not what the map adds.
- The notes were written by separate writers that read long stretches at once,
  not by the indexer. The request was offered as a step ("you may ask once");
  whether a model asks unprompted in a live respond call is not tested.
- Two conversations, one author. Answering model and first judge are the same
  model.
- One-source questions were not in this run, so "the map costs nothing where
  search already works" is still unchecked.
- All 150 held-out questions are now spent. The next claim needs new
  questions, and better a new conversation.

## 3. Data structures

### 3.1 Atoms

An atom is `(turn, start, end)` over the verbatim body, plus derived
fields. Because turns are immutable, atom ids never change.

| Field | Meaning |
|---|---|
| `id` | `t000123#a4`: turn id plus atom number, beside today's `#pK` parts |
| `turn`, `seq`, `role`, `created` | from the turn |
| `start`, `end`, `sha` | byte range in the body and the sha of the span |
| `kind` | prose, code, list, table, log, heading-led section |
| `tokens` | estimated per profile, as today |
| `values` | numbers, dates, identifiers, paths and names in it (the value lane's tokenizer) |
| `entities`, `thread` | from the index, when it has run |
| `seg` | segmentation version, so a better splitter can be added without breaking old ids |

Segmentation is deterministic and Markdown-aware:

- paragraphs split at blank lines; a heading or label opens the atom of
  the block after it;
- a code block is not cut at its blank lines, and is cut only above the
  maximum size;
- blocks under about 60 tokens join the next block; anything over about
  720 tokens (a list, table, log or code block included) is cut at a
  blank line, then a line end, then a sentence end. As built, a piece of
  a table does not repeat the header row, and an excerpt cut out of a
  code block says so in its header (the design's "each group shown with
  the header row" is not built).

The concatenation of a turn's atoms gives back the body byte for byte.
Unit tests check this on bodies with code fences, long paragraphs,
multibyte text, leading blank lines and bytes that are not text.

**Titles, the Obsidian part (optional, at index time).** Evergreen-note
practice gives every note a title that states its claim ("Packets are
sized to 60 to 70 percent of the window"), so the title works like an API
for the note. The indexer can write, per atom, a one-line claim and one
to three questions the atom answers (doc2query). These help exactly where
the real probes missed: paraphrases and restated facts. They are index
entries: searched, shown on catalog lines with the atom id, never shown as
facts and never sent in place of the atom. A cheap model writes them in
batches at index time; this costs model calls, so it is opt-in and is
kept only if fresh probes improve.

**Why not A-MEM's notes.** A-MEM (NeurIPS 2025) builds Zettelkasten-style
notes with model-written keywords, tags and descriptions, links them, and
lets new notes update old ones. That is the closest published system to
this idea, but its notes are rewritten memory, so errors compound and the
original wording can be lost. Here the atom text is always the verbatim
span; titles and links are derived data that can be rebuilt or thrown
away.

**Literal Obsidian export.** `cx export-obsidian <ws> <dir>` could write
one note per atom (or per exchange) with wikilinks along the graph below.
Obsidian's graph view then becomes a way to inspect what the engine
thinks is connected. It is a debugging view, not storage.

### 3.2 The link graph

Edges are typed, weighted, derived by code, and stored in SQLite
(`edges(src, dst, type, weight)`):

| Edge | Built from | Used for |
|---|---|---|
| `next` / `prev` | atom order in a turn | readability expansion |
| `answers` | a reply's atoms to the user turn it answers | header line, expansion |
| `same-value` | a rare value token shared by two atoms (a port, an identifier, a path), weighted by rarity | multi-hop lookups: "what port did we pick for the service that X" |
| `same-entity` | entities from the index | the same, when indexed |
| `supersedes` | the fact ledger | never send a stale value without the one that replaced it |
| `quotes` | an atom containing 40 or more characters of an earlier atom verbatim, or naming its id | provenance |
| `thread` | thread membership | levels of detail |

Uses: an **expansion lane** that spreads score from the top hits and the
recent exchanges one or two hops along `same-value`, `same-entity` and
`answers` (personalized PageRank, as in HippoRAG, at a few thousand nodes
costs milliseconds); **supersession closure** (a selected value always
travels with its newest replacement); and **navigation** for recall
("expand t000123#a4").

### 3.3 Levels of detail

| Level | Unit | Verbatim form | Card form |
|---|---|---|---|
| 0 | atom | the span | title, values |
| 1 | turn | the body | today's card |
| 2 | exchange | user turn plus reply | first line of the question, ids |
| 3 | block of up to 50 turns in a thread | all of it | today's digest line |
| 4 | conversation | everything | session.md |

A packet shows each region of history at one level, chosen by relevance
and budget: the focus and the latest exchanges at level 1 or 2, the hits
at level 0 with their headers, the exchanges around the hits as level 2
cards, other threads as level 3 digests. Every level has ids, so the model
can zoom in by recall. RAPTOR builds a similar tree from model-written
summaries and retrieves from any level; here the upper levels are only
navigation, and anything the model relies on arrives verbatim.

### 3.4 Indexes

- FTS5 over atoms with the porter stemmer (section 2); the turn index
  stays for exact phrases across paragraph breaks.
- Value index (value token to atoms), from the value lane's tokenizer.
- Entity index (as today, when the indexer has run).
- Optional atom titles and questions in their own FTS column.
- Optional embeddings over atoms, only after you approve a local model
  download.
- Usage counters per atom: times selected, cited, recalled (section 6).

## 4. The packet compiler

The compiler turns (message, workspace, route) into a packet in passes,
like a database planner turns a query into a plan:

1. **Classify** the request by rules first: look-up question, continue
   the work, change code, write, plan. The class sets the budget shape and
   the output contract.
2. **Budget.** The plan cap stays (60 to 70 percent of the window), but
   the target is much smaller and per class: about 8k to 16k tokens of
   history for a look-up, more for writing or code. Longer inputs cost
   more, answer later, and long contexts degrade even simple tasks
   ("lost in the middle", "context rot").
3. **Whole-history check.** If the whole history fits the target (or the
   cap, when the prefix is cached, see section 5), send it all, in order.
   The real-transcript run showed selection losing to the full transcript
   when everything fit.
4. **Candidates.** Lanes return scored atoms: exact phrase, full text over
   atoms, value, entity, graph expansion, thread recency, pinned and
   recalled ids.
5. **Fuse** with reciprocal rank fusion, keeping each lane's winner, as
   today.
6. **Expand for readability**, by deterministic rules: the header names
   the turn, its date and the first line of the question it answers;
   an atom that opens with a back reference ("This", "It", "So", "Also",
   a list continuation) brings its previous atom; adjacent atoms from one
   turn merge into one span under one header; supersession closure.
7. **Select under the budget.** Must-include first (session, decisions in
   force, the focus, the latest exchanges). Then greedy by marginal
   utility per token, with a redundancy penalty for atoms that repeat
   values or wording already chosen. A guarantee needs more than this: a
   monotone submodular objective, selection by gain per token, and a
   fallback to the best single item give at least 0.405 of the optimum
   under a token budget. A subtractive redundancy penalty breaks
   monotonicity, so redundancy has to be expressed as saturating coverage
   for the bound to apply.
8. **Levels of detail** for everything not selected, within a small fixed
   budget: digests for far regions, cards near the hits.
9. **Lay out for the cache and for attention**: see section 5.
10. **Size and route.** Estimate tokens, cost and time per candidate
    profile (section 5), and pick the cheapest profile whose measured
    success rate for this request class is above a threshold
    (`cx stats` already keeps the data).
11. **Explain.** `cx explain <ws> "<message>"` prints the plan: each
    lane's top candidates, what was chosen and why, tokens per section,
    the length of the cache-stable prefix, and estimated cost and time per
    profile. It is the main debugging tool, and the probes run on top of it.

### Output side

- **Output contract per class**: target length and format, claims cite
  atom ids, no restating of the packet. Output tokens cost several times
  input tokens and set most of the wall time.
- **Smaller packets for plan units**: each unit of a plan gets only the
  atoms its sub-question needs, so parallel units run on small, fast
  packets.
- **Small first, grow on demand** (a cascade): for look-ups, start with a
  small packet on a fast model; a `needs_recall` grows it. Whether the
  extra round trip pays for itself is measured, not assumed.

## 5. Cost, time and the prompt cache

Billed input is not the same as tokens sent. Providers charge cached
prefix tokens at a fraction of the normal price, so the order of the
packet can matter as much as its size. Two 2026 results bear on this. On
long agentic tasks, prompt caching cut cost by 27.8 to 81.4 percent
depending on model and strategy, and first-token time mostly improved but
regressed in some settings ("Don't Break the Cache"). In coding agents,
removing 38.4 percent of tool-output tokens raised the billed cost by 6.8
percent; the authors attribute that to changed agent trajectories and do
not show cache invalidation ("Token Reduction Is Not Cost Reduction"). So
measure the bill, not the tokens.

Layout, from most stable to least:

1. session and standing instructions;
2. decisions and facts in force;
3. **the append-only part**: the catalog and digests in id order. Each
   new turn only adds lines at the end, so each call's prefix extends the
   previous call's prefix and stays cached;
4. the request-specific spans, in chronological order, with headers;
5. the latest exchanges;
6. the focus message, last.

No run id, timestamp or other per-call text before the volatile part.
The brief now follows this order, with the run id and task at the end and
the message to answer last. That costs nothing. Two further steps are built
but off: `CatalogSpine` (the catalog lists every older turn) and
`recent_epoch` (the recent window starts at a multiple of N turns).

`cx cache` replays each user message and measures the prefix shared with the
previous brief, with no model call. On the three real conversations, in
selection mode, the old layout shared 0 percent. With both steps at an epoch
of 4 the median shared prefix was 37, 45 and 44 percent. If every call were
served that prefix at a tenth of the price, billed-equivalent input per
message would fall from 14,381 to 10,267, 18,741 to 11,632 and 25,483 to
18,312. With no hits it would rise to 15,296, 19,910 and 28,442. Probe
recovery did not change on any of the five sets.

Those savings are a ceiling. As documented in November 2026, Anthropic and
OpenAI's GPT-5.6 and later reuse prefixes only up to block or message
boundaries, and the brief is one text block. Claude Code shares a cache
across fresh runs only when its startup context matches, and each run here
has its own directory. So the two steps stay off until a real test shows
hits: send A+B then A+C, with A as part of one block and as its own block,
from one stable directory, and read the cache fields in `usage.json`.

**When the whole history fits**, sending it verbatim as an append-only
prefix may be both the best and the cheapest packet once it is cached,
since each call pays full price only for the new turns. Whether the CLIs
keep our prompt cached across separate headless runs is not known yet;
`usage.json` already records cache reads and writes per run, so this can
be measured on the first real calls.

Per-profile cost and time model, fitted from the runs table:

```text
cost = p_in * uncached_in + p_read * cached_in + p_write * newly_cached + p_out * out
time = a + b * uncached_in + c * out
```

`cx explain` prints both per candidate profile, and the router uses them.
On subscription CLIs the "price" is quota; the same model applies, with
the quota cost per token in place of dollars.

## 6. Recall as paging

A `needs_recall` is a page fault: the packet lacked something the model
needed. Treating it that way suggests three mechanisms:

- **Working set.** Atoms recalled or cited in the last few calls stay in
  the next packets, in the stable part (temporal locality, and cheap once
  cached).
- **Prefetch.** After a fault on an atom, its graph neighbors join the
  next packet (spatial locality).
- **Fault rate** per call is a health metric beside tokens per answered
  call. A rising fault rate means the selection is too stingy; zero faults
  with large packets means it is too generous.

Recall gains two verbs beside ids, search and parts: `expand` (neighbors
or the whole exchange of an atom) and `zoom out` (the digest of its
thread block).

## 7. Learning without a model in the loop

Signals the system already produces, or can for free:

| Signal | Label | Strength |
|---|---|---|
| Probes (offline) | the atom with the answer should be in the packet | strong, but written after the fact |
| `needs_recall` ids | those atoms should have been selected for that message | strong |
| Citations in the answer | those atoms were useful | medium |
| Selected, never cited, over many calls | probably low value for such requests | weak |
| User corrections ("no, it was 5433") | the old atom is stale | strong, rare |

Use them to tune lane weights, lane depth, the per-class budget and the
expansion rules, per workspace, by replaying past messages through the
compiler offline. The result is a small versioned config in the
workspace; selection stays deterministic. Every tuning is checked on
probes it did not see: the value lane was tuned on 14 synthetic probes
and added only one probe out of 30 on real ones.

## 8. Unusual ideas, collected

1. **A query planner and `EXPLAIN` for context** (section 4).
2. **Levels of detail like map tiles** (3.3): far history at low
   resolution, the relevant spot at full resolution, and zoom by recall.
3. **Recall as paging**: working set, prefetch, fault rate (section 6).
4. **The append-only history as the cache prefix** (section 5): while
   history fits, the full verbatim transcript may beat any selection on
   both quality and cost.
5. **Atom titles as APIs**, plus the questions each atom answers (3.1).
6. **The model's own citations and recall requests train the ranker**
   (section 7).
7. **Obsidian as a debugger** for the atom graph (3.1).

## 9. How to evaluate

- **Atom-level probes**: is the span holding the answer in the packet;
  tokens per recovered fact; recovery against budget at 2k, 4k, 8k, 16k
  and 32k tokens.
- **End to end on a sample**: a real model answers the probe questions;
  score the answers; record billed cost (cache included) and wall time
  from `usage.json`. Compare the full history (when it fits), today's
  turns, atoms with levels of detail, atoms with titles, and the CLI with
  its own compaction. This is the test that decides whether atoms are
  readable enough.
- **Harder questions**: probes written by you, without looking at the
  source, so they do not share its wording; and a LongMemEval or LoCoMo
  conversation.
- Report billed cost and time, not token counts.

## 10. Build plan

Each phase ends with a measurement that decides whether the next one is
worth building.

| Phase | Builds | Model calls | Kept if |
|---|---|---|---|
| 1 (built) | atom splitter, atom FTS with stemming, `#aK` recall ids, ranked atoms with headers in place of ranked turns, whole history when it fits, `cx probe -granularity` | none | atom probes at least match turn probes at a third of the tokens, on fresh probes. Result: missed at a third (45 against 48 of 60), met at the 8k default (52 against 48, for equal or fewer tokens); see 2b |
| 2 | cache-stable layout (run id moved), cost and time model in sizing, `cx explain`, cache reads measured | a few real runs | billed cost per answered message drops |
| 3 | link graph, expansion lane, readability rules, budgeted selection with redundancy penalty | none | multi-hop probes improve; nothing else drops |
| 4 | end-to-end evaluation; then atom titles and questions; learning from recall and citations; routing by cost | yes (ask first) | answers at least as good at lower cost or time |
| 5 | optional embedding lane; LongMemEval or LoCoMo | download (ask first) | beats phases 1 to 4 on paraphrased questions |

Phase 1 is small: the splitter, a table, one FTS index, a recall id form,
and a change to what the "relevant" selector emits. Everything after it
is optional and gated.

## 11. Risks

- **Readability.** An atom out of context can mislead. Headers, the
  question line and the back-reference rule are the mitigations, and only
  the end-to-end test shows whether they are enough.
- **Cache against selection.** Request-specific selection changes bytes
  in the prompt. Keeping it after the stable prefix limits the damage;
  billed cost is the measure.
- **CLI opacity.** Each vendor CLI adds its own system prompt and may or
  may not cache our text across runs. Measure per provider.
- **Over-engineering.** Plain bm25 over atoms already gets most of the
  gain in section 2. Each later phase has to beat phase 1 on probes it
  was not tuned on.
- **Index-time model output** (titles, questions, entities) can be wrong.
  It is never shown as fact, quotes are checked where there are quotes,
  and it can be rebuilt or dropped.

## 12. Prior art

| Work | Idea | What this design takes, and what it does differently |
|---|---|---|
| Dense X Retrieval (2023) | propositions as retrieval units | small units; but verbatim spans, not model-rewritten propositions |
| SeCom (2025) | conversational memory at segment granularity; compression as denoising | granularity matters; no compression here |
| A-MEM (NeurIPS 2025) | Zettelkasten notes with links that evolve | titles and links; never rewritten text |
| RAPTOR (2024) | a tree of summaries, retrieval at any level | levels of detail; upper levels only navigate |
| HippoRAG (2024) | knowledge graph plus personalized PageRank | the expansion lane over value and entity edges |
| Lost in the Middle (2023) | models use the start and end of long inputs best | focus last, stable material first |
| Don't Break the Cache (2026) | caching cut cost 27.8 to 81.4 percent, by model and strategy | the layout in section 5 |
| Token Reduction Is Not Cost Reduction (2026) | removing tokens can raise the bill | measure billed cost |
| TokenPilot (2026) | stable prefixes and eviction of expired segments, 56 to 87 percent cheaper | the append-only prefix; eviction by relevance |
| Lin and Bilmes (2011) | budgeted selection as a submodular problem | the selection step |

## Sources

- [Dense X Retrieval: What Retrieval Granularity Should We Use?](https://arxiv.org/abs/2312.06648) (Chen et al., 2023)
- [On Memory Construction and Retrieval for Personalized Conversational Agents](https://arxiv.org/abs/2502.05589) (Pan et al., 2025; the SeCom method)
- [A-MEM: Agentic Memory for LLM Agents](https://arxiv.org/abs/2502.12110) (Xu et al., NeurIPS 2025)
- [RAPTOR: Recursive Abstractive Processing for Tree-Organized Retrieval](https://arxiv.org/abs/2401.18059) (Sarthi et al., 2024)
- [HippoRAG: Neurobiologically Inspired Long-Term Memory for Large Language Models](https://arxiv.org/abs/2405.14831) (Gutiérrez et al., 2024)
- [Lost in the Middle: How Language Models Use Long Contexts](https://arxiv.org/abs/2307.03172) (Liu et al., 2023)
- [Context Rot](https://research.trychroma.com/context-rot) (Chroma, 2025)
- [Don't Break the Cache: An Evaluation of Prompt Caching for Long-Horizon Agentic Tasks](https://arxiv.org/abs/2601.06007) (Lumer et al., January 2026)
- [Token Reduction Is Not Cost Reduction](https://arxiv.org/abs/2607.12161) (Weinberger and Hozez, July 2026)
- [TokenPilot: Cache-Efficient Context Management for LLM Agents](https://arxiv.org/abs/2606.17016) (Xu et al., June 2026)
- [Document Expansion by Query Prediction](https://arxiv.org/abs/1904.08375) (Nogueira et al., 2019; doc2query)
- Andy Matuschak, "Evergreen note titles are like APIs", in his [working notes](https://notes.andymatuschak.org/)
- [A Class of Submodular Functions for Document Summarization](https://aclanthology.org/P11-1052/) (Lin and Bilmes, ACL 2011)
