## BOOTSTRAP — what to do after extracting this bundle

This document walks you through the steps to take after extracting the bundle into your freshly-cloned, empty `fuzzymatch` repository. Follow them in order.

The bundle is designed to **follow GSD's natural workflow** — including its research phase. Our `docs/requirements.md` and `docs/prior-art-research.md` serve as input material that GSD's research agents will consult, not as a substitute for GSD's research process.

---

## 0. Where you are now

You have:

- An empty (or almost-empty) GitHub repository for `github.com/axonops/fuzzymatch`, cloned locally.
- This bundle extracted at the repo root.
- The GSD plugin (`get-shit-done-cc`) installed in Claude Code.

You should see this structure in the repo root:

    .
    ├── CLAUDE.md
    ├── LICENSE
    ├── NOTICE
    ├── README.md
    ├── CHANGELOG.md
    ├── BOOTSTRAP.md
    ├── gsd-agent-skills.json
    ├── .gitignore
    ├── claude/                     (RENAME to .claude/ — see step 1)
    │   ├── skills/
    │   │   ├── algorithm-correctness-standards/SKILL.md
    │   │   ├── algorithm-licensing-standards/SKILL.md
    │   │   ├── commit-standards/SKILL.md
    │   │   ├── determinism-standards/SKILL.md
    │   │   ├── documentation-standards/SKILL.md
    │   │   ├── fuzzymatch-review-protocol/SKILL.md
    │   │   ├── go-coding-standards/SKILL.md
    │   │   ├── go-testing-standards/SKILL.md
    │   │   ├── issue-standards/SKILL.md
    │   │   ├── performance-standards/SKILL.md
    │   │   └── research-guidance/SKILL.md
    │   └── agents/
    │       └── ... (17 files)
    └── docs/
        ├── requirements.md         (authoritative spec, 117 KB, 23 algorithms)
        └── prior-art-research.md   (Go ecosystem survey + algorithm taxonomy)

---

## 1. Rename `claude/` to `.claude/`

The bundle uses `claude/` (no leading dot) because some zip tools mangle dot-directories. Claude Code expects `.claude/`. Rename now:

    mv claude .claude

Verify:

    ls -la .claude/
    ls .claude/skills/ | wc -l           # should show 11
    find .claude/skills -name SKILL.md | wc -l   # should show 11
    ls .claude/agents/ | wc -l            # should show 17

If any of these do not match, stop and investigate before continuing.

---

## 2. Initial commit and push

Get the repo into a clean state before letting GSD touch it:

    git add .
    git commit -m "chore: bootstrap repository"
    git push origin main

If your default branch is `master`, adjust. If the GitHub repo was created with an initial commit (LICENSE/README from the GitHub UI), resolve duplicates first.

---

## 3. Launch Claude Code with GSD

In the repo root:

    claude --dangerously-skip-permissions

GSD's README recommends `--dangerously-skip-permissions` for friction-free automation.

---

## 4. Map the existing code

Even though we have no production Go code yet, running `/gsd-map-codebase` makes GSD aware of `CLAUDE.md`, the skills, the agents, the docs, and the licence — so the questions in `/gsd-new-project` are sharper.

    /gsd-map-codebase

This produces `.planning/codebase/STACK.md`, `.planning/codebase/CONVENTIONS.md`, etc. Briefly review the output — especially `CONVENTIONS.md` — to verify GSD picked up the Apache-2.0 licence, the 23-algorithm intent, the review-agent gates, and the presence of `docs/requirements.md` and `docs/prior-art-research.md`.

---

## 5. Run `/gsd-new-project` — follow the natural GSD flow

This is GSD's full Questions → Research → Requirements → Roadmap cycle. Let it run naturally.

    /gsd-new-project

### During the questions phase

Answer GSD's questions honestly. When it asks about scope, dependencies, or target users, you can reference `docs/requirements.md` as the authoritative spec — but answer the questions yourself rather than pointing GSD at the file. GSD's questions exist to extract decisions and constraints that may not be in the spec.

### During the research phase

GSD will spawn `gsd-project-researcher` agents in parallel (typically 4 — for STACK, FEATURES, ARCHITECTURE, PITFALLS). Because of the `agent_skills` config you'll apply in step 6, these researchers will read:

- `research-guidance/SKILL.md` (tells them what to focus on)
- `algorithm-licensing-standards`, `algorithm-correctness-standards`, `performance-standards`, `determinism-standards`, `go-coding-standards`

And they can read the input docs already in your repo:

- `docs/requirements.md` — the authoritative spec
- `docs/prior-art-research.md` — the Go ecosystem survey

The researchers should converge on findings that align with `docs/requirements.md` (both work from the same source material). Where they diverge → review the divergence; it might be GSD seeing something missed, or GSD being thinner than our spec.

**Caveat:** if you have not yet applied the `agent_skills` config (step 6) when `/gsd-new-project` runs, the project researchers will run without skill injection. That is still fine — they'll have access to the source docs in the repo. But the research will be richer once skills are wired in. For the most thorough first research pass, apply step 6 *before* step 5 (you can pre-create `.planning/config.json` with just the `agent_skills` block; GSD will preserve it when it adds its own fields).

### During requirements extraction

GSD's `REQUIREMENTS.md` is intended as the high-level scope (v1, v2, out-of-scope). Our `docs/requirements.md` is the deep technical spec. The two should coexist; `REQUIREMENTS.md` should reference `docs/requirements.md` for technical detail rather than duplicating 117 KB of it. If GSD's REQUIREMENTS.md tries to absorb the full spec, edit it after generation to reference instead.

### During roadmap creation

GSD's `ROADMAP.md` should produce roughly eight phases corresponding to `v0.1.0 → v0.2.0 → v0.3.0 → v0.4.0 → v0.5.0 → v0.6.0 → integration → v1.0.0` per `docs/requirements.md` §19. If GSD's phasing differs:

- Accept GSD's structure if you prefer it
- Edit `.planning/ROADMAP.md` to match the eight phases
- Or re-run with explicit phasing guidance

### Review and approve

GSD will ask you to confirm PROJECT.md, REQUIREMENTS.md, and ROADMAP.md before proceeding. Read them carefully. This is where you catch divergences early.

---

## 6. Configure `agent_skills` injection

After `/gsd-new-project` completes, `.planning/config.json` exists. Merge in the `agent_skills` block from `gsd-agent-skills.json`.

**Two options:**

### Option A — `gsd-tools.cjs config-set` (recommended)

Run one command per agent type. The exact commands are in `gsd-agent-skills.json` — copy the array values into the `config-set` argument. Example for `gsd-project-researcher`:

    node ~/.claude/get-shit-done/bin/gsd-tools.cjs config-set agent_skills.gsd-project-researcher '[".claude/skills/research-guidance", ".claude/skills/algorithm-licensing-standards", ".claude/skills/algorithm-correctness-standards", ".claude/skills/performance-standards", ".claude/skills/determinism-standards", ".claude/skills/go-coding-standards"]'

Repeat for: `gsd-research-synthesizer`, `gsd-phase-researcher`, `gsd-planner`, `gsd-executor`, `gsd-verifier`, `gsd-plan-checker`, `gsd-debugger`. The values are in `gsd-agent-skills.json`.

### Option B — edit `.planning/config.json` by hand

Open `.planning/config.json`. Find or add the `agent_skills` key at the top level. Paste the content from `gsd-agent-skills.json` (everything inside the `agent_skills` block, without the `_comment` field). Save.

### Verify

    cat .planning/config.json | python3 -m json.tool | grep -A 80 agent_skills

You should see eight agent types each mapped to a list of skill directory paths.

### Commit

    git add .planning/
    git commit -m "chore(gsd): configure agent_skills for fuzzymatch domain"

You can now safely delete `gsd-agent-skills.json` from the repo root:

    rm gsd-agent-skills.json
    git add -A
    git commit -m "chore: remove agent_skills snippet"

---

## 7. Sanity-check the integration

Verify the agents and skills are correctly wired:

    find .claude/skills -name SKILL.md | wc -l    # 11
    ls .claude/agents/*.md | wc -l                 # 17

Verify the JSON config:

    python3 -c "import json; c=json.load(open('.planning/config.json')); print(len(c.get('agent_skills',{})))"
    # should show 8

Verify the key wiring:

    python3 -c "import json; c=json.load(open('.planning/config.json')); print('fuzzymatch-review-protocol' in str(c['agent_skills'].get('gsd-verifier',[])))"
    # should show True

    python3 -c "import json; c=json.load(open('.planning/config.json')); print('research-guidance' in str(c['agent_skills'].get('gsd-project-researcher',[])))"
    # should show True

---

## 8. Begin the GSD loop

You are now ready to build. The loop:

    /gsd-discuss-phase 1      capture implementation decisions for phase 1
    /gsd-plan-phase 1         gsd-phase-researcher → gsd-planner → gsd-plan-checker
    /gsd-execute-phase 1      gsd-executor builds in parallel waves
    /gsd-verify-work 1        gsd-verifier invokes our review agents
    /gsd-ship 1               PR or merge

Repeat for phase 2, 3, ... 8.

### Watch for on the first phase

1. **During `/gsd-plan-phase 1`:** `gsd-phase-researcher` should produce `.planning/phases/01-*/01-RESEARCH.md` informed by `research-guidance`, citing primary academic sources and respecting the patent screen.

2. **During `/gsd-execute-phase 1`:** `gsd-executor` should produce code with Apache-2.0 file headers, the algorithm citation block, reference vectors in tests, BDD scenarios.

3. **During `/gsd-verify-work 1`:** `gsd-verifier` should invoke our review agents via `Task` calls. Specifically watch for:

   - `algorithm-correctness-reviewer` (for any algorithm implementation)
   - `algorithm-performance-reviewer` (benchmark verification)
   - `determinism-reviewer` (output-stability verification)
   - `algorithm-licensing-reviewer` (any new algorithm)
   - `bdd-scenario-reviewer` (BDD coverage)
   - `code-reviewer` (general)
   - `go-quality` (automated quality)

If verification advances without invoking these agents, the integration is misconfigured. Check:

- `fuzzymatch-review-protocol` is in `agent_skills.gsd-verifier`
- `.claude/skills/fuzzymatch-review-protocol/SKILL.md` exists and is readable
- `CLAUDE.md` is at the repo root, not nested under `.claude/`

---

## 9. About the issue-tracking files

Two skills and two agents in this bundle were originally authored for a GitHub-issues workflow we replaced with GSD:

- `.claude/skills/issue-standards/SKILL.md`
- `.claude/agents/issue-writer.md`
- `.claude/agents/issue-closer.md`

These are harmless and can stay. They become useful again if GitHub Issues are needed for external bug reports after release. If you'd rather not keep them:

    rm -rf .claude/skills/issue-standards
    rm .claude/agents/issue-writer.md .claude/agents/issue-closer.md
    git add -A
    git commit -m "chore: remove issue-tracking workflow artefacts"

---

## 10. Troubleshooting

**Research agents do not read `research-guidance`:**

- Check `.planning/config.json` has `agent_skills.gsd-project-researcher` containing `.claude/skills/research-guidance`.
- Verify the path: `.claude/skills/research-guidance/SKILL.md` (case-sensitive).
- Run `node ~/.claude/get-shit-done/bin/gsd-tools.cjs agent-skills gsd-project-researcher` and check the output lists the skill.

**Review agents do not fire during `/gsd-verify-work`:**

- Check `agent_skills.gsd-verifier` includes `fuzzymatch-review-protocol`.
- Re-read `.claude/skills/fuzzymatch-review-protocol/SKILL.md`.
- Confirm `CLAUDE.md` is at the repo root.

**`/gsd-new-project` ignored `docs/requirements.md`:**

- During the interactive questions, mention the file explicitly: "The authoritative spec is `docs/requirements.md`. Read it during the research phase."
- After GSD completes, edit `.planning/REQUIREMENTS.md` to reference `docs/requirements.md` as the deep-dive.

**GSD's roadmap has different phasing than `docs/requirements.md` §19:**

- Accept it if you prefer GSD's structure.
- Or edit `.planning/ROADMAP.md` directly.
- Or re-run `/gsd-new-project` and steer the roadmap.

---

## After bootstrap

Once everything works, you can delete this `BOOTSTRAP.md` file. Or keep it as a reference for the next contributor.

Good luck.
