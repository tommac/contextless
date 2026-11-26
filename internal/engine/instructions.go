package engine

import (
	"fmt"
	"strings"
)

func respondInstructions(runID string) string {
	return respondTextMax(runID, true, defaultMaxUnitsHint)
}

// respondText is the responder's task; without allowPlan it offers no
// plan (a plan the engine refused, or a single-call trial).
func respondText(runID string, allowPlan bool) string {
	return respondTextMax(runID, allowPlan, defaultMaxUnitsHint)
}

func respondTextMax(runID string, allowPlan bool, maxUnits int) string {
	if maxUnits <= 0 {
		maxUnits = defaultMaxUnitsHint
	}
	plan := ""
	if allowPlan {
		plan = fmt.Sprintf(`
- plan: the request is large or has parts better done separately and in parallel (research on the web, a long piece of code, reading very long material, several independent sections), or it asks to change an attached repository. Return 2 to %[1]d `+"`units`"+` (one is enough for a coder unit that changes a repository). Each unit has an id (u1, u2, ...), a role (responder for reasoning and writing, coder for code, long_reader for very long material, researcher for anything that needs the web, editor for polishing prose, reviewer for checking work), a title, self-contained instructions, `+"`needs`"+` (turn ids the unit must see verbatim), `+"`repo`"+` (the name of an attached repository a coder unit changes, else empty), and `+"`depends_on`"+` (ids of earlier units whose results this unit needs: it runs after them, with their records). Units with nothing to wait for run in parallel, at most %[2]d at a time, on different models, and an assembler merges their records. Put the choices every unit must follow (names, interfaces, formats, units of measure, what is in and out of scope) in `+"`shared_decisions`"+`; every unit gets them. A plan of one unit, or one whose units would only run one after another, is not run: answer such a request yourself, with partial if it will not fit. A coder unit with a repo works in a fresh checkout on its own branch; its changes are checked against the allowed paths and verified, and the user decides whether to merge; coder units on the same repository run one after another, each on the previous one's commit. Leave `+"`answer`"+` empty.`, maxUnits, maxParallelUnits)
	}
	text := fmt.Sprintf(`Answer the message in the last section ("The message to answer") as the assistant in this conversation. The conversation records, decisions and facts in this packet are verbatim and authoritative.

Choose one status:

- completed: `+"`answer`"+` holds your full reply to the user, in Markdown.
- needs_recall: you need turns, paragraphs or parts that are not in this packet word for word. Put their ids in `+"`recall_ids`"+` (at most 20), search words in `+"`recall_queries`"+`, or typed operations in `+"`recall_ops`"+` (exchange, open_block, timeline, grep, notes), and leave `+"`answer`"+` empty. The engine will run you again with that text included verbatim.%s
- partial: you began a long answer that will not fit in this call. Put what is done in `+"`answer`"+`, list what remains in `+"`units_remaining`"+`, and say exactly where to resume in `+"`next_step`"+`.
- blocked: you need something only the user can give. `+"`answer`"+` holds your question.
`, plan)
	if allowPlan {
		text += "\nPrefer completed. Use plan only when splitting clearly helps the user.\n"
	}
	return text + fmt.Sprintf(`Set run_id to "%s". Fill every field you do not use with an empty string or an empty array.`, runID)
}

// defaultMaxUnitsHint is the fallback unit cap quoted in respond instructions.
const defaultMaxUnitsHint = 8

func unitInstructions(runID string, u unitSpec, total int, shared bool) string {
	var context []string
	if len(u.DependsOn) > 0 {
		context = append(context, fmt.Sprintf("This unit depends on %s: their records are under \"Supplied records\", checked by the engine like yours will be.", strings.Join(u.DependsOn, ", ")))
	}
	if shared {
		context = append(context, "The plan's shared decisions are under \"Supplied records\": follow them, and if one cannot be followed, say so in conflicts.")
	}
	more := ""
	if len(context) > 0 {
		more = "\n" + strings.Join(context, " ") + "\n"
	}
	return fmt.Sprintf(`You are unit %s of %d in a plan answering the message in the last section. Do only this unit.

Unit: %s
Role: %s
Instructions:
%s
`+more+`
%s

Status completed; no_information if this packet and your tools hold nothing for this unit's task (say in notes what you looked for, and do not invent an answer); partial (with units_remaining and next_step) if it will not fit; needs_recall (with recall_ids or recall_queries) if you need omitted text verbatim; blocked if the unit cannot be done (say why in answer).
Set run_id to "%s". Fill every field you do not use with an empty string or an empty array.`,
		u.ID, total, u.Title, u.Role, strings.TrimSpace(u.Instructions), unitRecordGuide, runID)
}

func assembleInstructions(runID string, n int) string {
	return fmt.Sprintf(`%d units worked on the message in the last section. Their records are under "Supplied records": each unit's output verbatim, then the claims it relies on with what the engine found when it checked each quote ("quote found in tNNNNNN#aK" means the quote is in that stored text; "quote not found" means it is not, so the claim counts as inferred; "from a web source the engine did not check" means just that), then its assumptions, open questions, conflicts and notes. A unit that found no information is listed without content.

Before the records, the engine's comparison of them (computed by code, not by a model) lists claims of different units that cite the same stored text, marking those that state different values, assumptions of different units that may be about the same choice, and the units' open questions and conflicts.

Write the single final reply to the user from them. Keep every substantive result; merge overlaps. Rely on claims whose quote was found; do not present a claim whose quote was not found, or that has no evidence, as established fact. When units disagree, or made different assumptions about the same thing, settle it from the cited text if you can, else say so and say which you trust and why; never pick one silently. Carry the open questions and conflicts that matter to the user into the reply. Do not mention the units or this process unless that helps the user.

Account for every claim id in the records (u1.c1, u2.c3, ...): put the ids your reply relies on in `+"`claims_used`"+`, and every other one in `+"`claims_omitted`"+` with the reason you left it out (a duplicate of another claim, contradicted by the cited text, not needed for this reply, unsupported). The engine checks that each claim is in one list or the other. Do not write claim ids in the reply itself.

Status completed with the reply in `+"`answer`"+` (or partial with next_step if the reply will not fit, or needs_recall if you need omitted text verbatim).
Set run_id to "%s". Fill every field you do not use with an empty string or an empty array.`, n, runID)
}

func continueInstructions(runID string, nextStep string, remaining []string) string {
	rem := "(none listed)"
	if len(remaining) > 0 {
		rem = "- " + strings.Join(remaining, "\n- ")
	}
	return fmt.Sprintf(`A previous run began answering the message in the last section and stopped at a clean boundary. Its work so far is under "Supplied records", verbatim.

Continue from exactly where it stopped. Do not repeat anything already written. Your `+"`answer`"+` holds only the new part, which will be appended to the work so far.

Where to resume: %s

What remains:
%s

Status completed when this finishes the reply, or partial again (with units_remaining and next_step) if it still will not fit.
Set run_id to "%s". Fill every field you do not use with an empty string or an empty array.`, nextStep, rem, runID)
}

func chunkInstructions(runID string, part, total int, message string) string {
	return fmt.Sprintf(`The message in the last section is too large for one call, so it was split into %[2]d parts at paragraph boundaries; this packet holds part %[1]d verbatim. If the request itself is not in this part, it is in "Supplied records" (the opening of the message, verbatim).

Do the user's request for this part only: produce everything from this part that the final reply needs (findings, extracted items, answers, code). Do not summarize the rest of the message; other runs handle the other parts and an assembler merges them. In evidence, cite this message as %[4]s and quote from this part.

%[5]s

Status completed, or no_information if this part holds nothing for the request.
Set run_id to "%[3]s". Fill every field you do not use with an empty string or an empty array.`, part, total, runID, message, unitRecordGuide)
}

func indexInstructions(runID string, ids []string, withIdeas bool) string {
	ideas := ""
	if withIdeas {
		ideas = ideaInstructions
	}
	return fmt.Sprintf(`Index these turns for the conversation's catalog: %s. They are under "Conversation records" and "Turns to index", verbatim.

For each of these turns write one card:
- summary: one or two plain sentences saying what the turn is about, so it can be found later. It is a catalog line, never a replacement for the turn.
- threads: thread ids from "Existing threads", or keys of new threads you declare in new_threads as "new:<slug>". A thread is a topic or workstream that may recur.
- mentions: entity ids from "Existing entities", or keys of new entities you declare in new_entities as "new:<slug>" (people, projects, files, products, organizations, places, concepts that matter in this conversation).
- keywords: up to eight search words.

Also record:
- facts: things the user stated or that the conversation established, each with source (the turn id) and quote copied exactly, character for character, from that turn. A quote not found verbatim is discarded. When a fact changes a value listed under "Facts in force" (a new port, date, price, name or time), put the old fact's id in supersedes; the old fact is kept, marked replaced.
- decisions: things the user decided or asked to be done a certain way from now on, with source and an exact quote; supersedes lists decision ids from "Decisions in force" that a new decision replaces.
- thread_status: set a thread to resolved when these turns close it.

%sDo not invent facts the turns do not support. Status completed.
Set run_id to "%s". Fill every field you do not use with an empty string or an empty array.`, strings.Join(ids, ", "), ideas, runID)
}

// ideaInstructions asks the indexer for idea notes: the map of the
// conversation a later reader chooses its paragraphs from.
const ideaInstructions = `- ideas: notes for the map of the conversation, which a later reader uses to decide which paragraphs to read. Write one note for each distinct thing these turns say that someone may later need to find: a decision and its reason, a requirement, a fact or value, a result of something tried, a problem and its cause, a plan, a change to something said earlier, an open question. A long turn usually holds several; a turn of small talk holds none. For each:
  - title: one plain sentence that states the idea itself with its specifics (the names, numbers and identifiers that matter), understandable without the conversation. Not "discussion of the cache" but what was found or decided about it.
  - kind: one of decision, requirement, fact, result, problem, plan, change, question, attempt.
  - topics: one to three short subject names, reused from "Existing note topics" whenever one fits, so that notes on one subject are filed together.
  - evidence: one or more entries, each with turn (the turn id) and quote copied exactly, character for character, from that turn: the sentence or phrase that states the idea. A note whose quotes are not found verbatim is discarded, so copy, never paraphrase. When an idea is stated in two places (asked for in one turn, confirmed in another), give both.

`
