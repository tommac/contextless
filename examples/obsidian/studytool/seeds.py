"""Calibration seeds: code-built designs with one known defect each.

Every seed targets one metric with an expected score of at most 2:
  generic           -> originality (a familiar pattern, no distinctive mechanism)
  breaks-constraint -> feasibility (violates the problem's stated resources)
  weak-test         -> test (the comparison cannot disfavor the mechanism)

Seeds are deterministic so reviewers of any round face identical traps.
"""

GENERIC = {
    "A": "**Watch and Try.** Design 1: the electrician demonstrates each repair while visitors watch. Design 2: visitors hold parts and help. Develop watching: everyone gathers around the bench, the electrician explains each step aloud, and visitors may ask questions. Example: a kettle gets its fuse replaced while twelve people watch. Test: after the session, ask visitors whether they feel more confident about fixing things.",
    "B": "**Take Turns.** Design 1: the musicians agree to alternate every hour. Design 2: both play quietly all night. Develop alternating: North plays 7 to 8, South plays 8 to 9, and so on, so each end gets the crowd in turn. Example: at 8 o'clock North stops and South starts a song. Test: at the end of the night, ask both musicians whether they are happy with how the evening went.",
    "C": "**First In First Out.** Design 1: play tapes in arrival order each night. Design 2: draw tapes from a hat. Develop arrival order: number each tape as it arrives and air the oldest first until three hours are full. Example: tape 14 airs before tape 15. Test: after a week, send residents a survey asking whether they are satisfied with the station.",
    "D": "**A Fitting Farewell.** Design 1: hold a ceremony before demolition. Design 2: install a plaque on the new crossing. Develop the ceremony: invite residents to walk the bridge one last time, share memories, and take photos. Example: a Sunday walk with speeches at the midpoint. Test: count how many people attend and ask them afterwards whether the event felt meaningful.",
}

BREAKS_CONSTRAINT = {
    "A": "**Rotating Tutors.** Three hired electricians each lead a four-person workbench, and an app matches arrivals to benches by fault type; visitors pre-register online the week before. Each visitor disassembles a device with their tutor. Example: the kettle bench rewires three kettles together. Test: quiz visitors on an unseen toaster; bench participants should diagnose more faults than drop-ins.",
    "B": "**Shared Mixer.** Both musicians plug into one rented amplifier with a mixing board placed at the market center, and a hired sound engineer rides the levels all night, texting each performer when to swell or soften. Example: at 8:20 the engineer texts North to go forte. Test: compare tips on engineer nights versus unmixed nights; the mixer should keep both cases fuller.",
    "C": "**Smart Scheduler.** An app transcribes every tape on arrival, and a paid overnight DJ assembles the three hours live from the transcript summaries, texting donors whose tapes aired. Example: Ana gets a 1am text that her story played. Test: count distinct voices per week under the DJ versus arrival order; the DJ should air more voices.",
    "D": "**Daily Meetings.** The team holds a public meeting every week instead of one per season, hires a facilitator and a film crew, and buys out a billboard campaign, spending about ten times the modest budget to make sure every voice is heard. Example: weekly hall sessions with livestream. Test: compare attendance at the first and last meetings; attendance should grow.",
}

WEAK_TEST = {
    "A": "**Fix-it Passport.** Each visitor gets a passport stamped for every fault type they touch: power, heating, moving parts. The electrician demos one, then visitors pair up on the same fault. Example: two visitors test a fuse on a second kettle after the demo. Test: hold the cafe for two months and see whether attendance grows and people say nice things in the guestbook.",
    "B": "**Mirror Sets.** Both musicians play the same twelve-song list in the same order, starting on the hour, so neither end sounds livelier than the other. Example: both open with the same waltz at 7. Test: run it for a few nights and see whether the market feels pleasant and whether the musicians want to keep doing it.",
    "C": "**Star Chart.** A corkboard shows every tape as a star; tapes with fewest stars air first, and each airing adds a star. Example: a tape with zero stars airs before one with three. Test: watch the board over a month and check that it fills up nicely with stars across most tapes.",
    "D": "**Memory Walk.** Monthly guided walks collect stories at the bridge; the best are printed in a booklet placed in the new crossing's railing kiosk. Example: a ferry-shift story becomes a railing panel. Test: print the booklet and see whether people pick it up and whether the team feels proud of it.",
}


def make_seeds():
    seeds = []
    for defect, table in (("generic", GENERIC), ("breaks-constraint", BREAKS_CONSTRAINT),
                          ("weak-test", WEAK_TEST)):
        for problem in sorted(table):
            seeds.append({"problem": problem, "defect": defect, "answer": table[problem]})
    return seeds


if __name__ == "__main__":
    import json
    print(json.dumps(make_seeds(), ensure_ascii=False, indent=2))
