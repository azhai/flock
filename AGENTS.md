# AGENTS.md — AI Programming Guidance

This file governs how AI coding agents operate in this repository. It is a
synthesis of three proven, openly published approaches, each contributing a
distinct, complementary discipline:

| Source | What it contributes | Core idea |
| --- | --- | --- |
| [OthmanAdi/planning-with-files](https://github.com/OthmanAdi/planning-with-files) | **State persistence** | Treat the filesystem as durable memory; plan in files so goals survive `/clear`, crashes, and context compaction. |
| [DietrichGebert/ponytail](https://github.com/DietrichGebert/ponytail) | **Output discipline** | Write the minimum that works; reuse before you write; never cut safety. |
| [deepseek-ai/deepseek-harness](https://github.com/deepseek-ai/deepseek-harness) | **Architecture & verification** | Build everything as a swappable plugin ("capability seam"); verify before you declare done; record decisions as ADRs. |

These are not in tension. A minimal solution that is planned on disk, composed
from plugins, and proven by tests is the target state for every task.

---

## 0. Core Operating Principles

Hold these six rules in mind on every turn. They are elaborated in the sections below.

1. **Disk is memory; context is RAM.** Anything important is written to a file, not only kept in the conversation.
2. **Plan before you build.** No task of 3+ steps or 5+ tool calls starts without a plan file.
3. **Write the minimum that works.** Reuse before you write; the best code is the code you never shipped.
4. **Compose with plugins, not edits.** Behavior changes by swapping a provider, not by rewriting call sites.
5. **Verify before you claim done.** Run the repo's gates; never repeat a failure you already logged.
6. **Safety is non-negotiable.** Trust-boundary validation, data-loss handling, security, and accessibility are never on the chopping block.

---

## 1. Plan in Files, Not in Context

Adopt the Manus-style context-engineering model from *planning-with-files*:
the **context window is volatile and limited (RAM)**, the **filesystem is
durable and effectively unlimited (disk)**. Persist execution state to files so
a `/clear`, a crash, or a compaction cannot lose the plan.

### 1.1 The 3-file pattern

For any multi-step task, maintain three files (templates in
`.ai-memory/planning-template.md`):

| File | Purpose |
| --- | --- |
| `task_plan.md` | Phases, checkboxes, and the resume point after a reset. |
| `findings.md` | Research, decisions, and things learned along the way. |
| `progress.md` | A session log: commands run, outputs, test results, errors. |

### 1.2 When to create them (ladder)

Stop at the first applicable rung:

1. Task needs 3+ steps or 5+ tool calls? → **create the three files first.**
2. Learned something new? → **append it to `findings.md`.**
3. Did something? → **log it in `progress.md`.**
4. A phase finished? → **check it off in `task_plan.md`.**
5. Context died (`/clear`, crash, compaction)? → **re-read the plan from disk and resume.**
6. Every phase complete? → **only then is the task done.**

### 1.3 Key rules

- **Create the plan first.** Never start a non-trivial task without `task_plan.md`.
- **The 2-action rule.** After every ~2 read/search/browse operations, save what you learned to `findings.md`.
- **Log ALL errors.** Failed commands go in `progress.md`; they prevent repeating the same mistake.
- **Never repeat failures.** Track each attempt; if an approach failed, mutate the approach before retrying.

### 1.4 Plan recitation and recovery

At the start of each turn, re-read the active plan before making a decision.
If the conversation was reset, the plan files on disk are the source of truth —
recover from them without asking the user to restate the goal.

---

## 2. Write the Minimum That Works (Ponytail)

Be the "lazy senior dev": ruthless about avoiding unnecessary code, meticulous
about correctness. Apply the **reuse ladder** *after* you understand the problem,
never instead of understanding it — lazy about the solution, never about reading.

### 2.1 The reuse ladder

Before writing code, stop at the first rung that holds:

1. **Does it need to exist?** — No → skip it (YAGNI).
2. **Already in this codebase?** — Yes → reuse it, don't rewrite.
3. **Stdlib does it?** — Yes → use the standard library.
4. **Native platform feature?** — Yes (e.g. `<input type="date">`, `fetch`) → use it.
5. **Installed dependency?** — Yes → use it; don't reinvent it.
6. **One line?** — Yes → write one line.
7. **Only then** → write the minimum that actually works.

### 2.2 Safety is non-negotiable

The ladder compresses *code*, never *correctness*. Never cut:

- **Trust-boundary validation** — sanitize and verify untrusted input.
- **Data-loss handling** — guard destructive operations; confirm before irreversible writes.
- **Security** — no secrets in code, no unsafe eval, no exposed surfaces.
- **Accessibility** — semantic markup, labels, and contrast are part of "working."

### 2.3 The `ponytail:` marker convention

When you deliberately cut a real corner with a known ceiling, mark it inline so
it is greppable and reviewable:

```ts
// ponytail: browser has <input type="date">; skip the flatpickr wrapper
```

Use this only for trivial, bounded simplifications — not as a license to ship
unsafe code. Periodically run a debt review (`grep` for `ponytail:`) to confirm
the ceilings are still acceptable.

Full ladder and marker rules: `.ai-memory/ponytail-ladder.md`.

---

## 3. Build Everything as a Plugin (Capability Seams)

From *deepseek-harness*: **everything is a plugin.** Behavior is composed from
swappable capabilities, not hardcoded decisions. This keeps the runtime closure
tight and makes the system testable and replaceable.

### 3.1 The Capability Seam

A *seam* is a swappable capability defined by three roles:

| Role | Responsibility | Example |
| --- | --- | --- |
| **Service Definition** | Declares the interface (the seam). | `ctx.llm`, `ctx.fs`, `ctx.shell` |
| **Service Provider** | Implements the interface. | `llm-deepseek`, `fs-local`, `fs-sandbox` |
| **Consumer** | Uses the capability, often a tool. | `tool-bash` consuming the shell capability |

Change behavior by swapping a **provider** — e.g. move from a local filesystem to
a sandboxed one — without touching the **consumer** call sites.

### 3.2 Rules

- **Define the seam before the implementation.** Interface first; providers are interchangeable.
- **Keep the runtime closure closed.** Don't let capability providers pull in unbounded transitive deps.
- **Compose via an event bus / plugin host** (Cordis-style) rather than direct wiring.
- **Prefer configuration over forking.** A new behavior is a new plugin, not a branch in core logic.

Capability-seam pattern and provider checklist: `.ai-memory/capability-seams.md`.

---

## 4. Verify Before You Claim Done

From *deepseek-harness*: an agent may not declare a task complete until it has
passed the repo's quality gates. Guessing "it probably works" is not verification.

### 4.1 Run the repo's gates

Use the commands the project already defines (see `AGENTS.md` / `package.json`
scripts). Typical gates:

```bash
lint            # static analysis / style
lint:fix        # auto-fix what is safe
test            # unit + integration
test:coverage   # coverage-gated run
typecheck       # type equivalence / build
knip            # dead-code / unused-export detection
hygiene         # cross-cutting checks (i18n keys, final newline, etc.)
```

If the repo defines its own gate command (e.g. `run-gates`), run **that**, not a
hand-picked subset.

### 4.2 Discipline

- **Re-run after every change.** A green run before an edit proves nothing after it.
- **Never skip verification** to "save time." An unverified claim is a bug waiting to ship.
- **Log the run** in `progress.md` with the command and its exit status.
- **Measure when possible.** Track LOC, tokens, cost, time, and safety impact so improvements are observable (ponytail's benchmark discipline).

---

## 5. Decide With ADRs

Significant architectural or behavioral choices are recorded as **Architecture
Decision Records** so future agents (and humans) understand *why*, not just *what*.

### 5.1 When to write an ADR

Write one when you introduce a **capability seam**, choose an **LLM/provider
adapter**, change a **public interface**, or make a trade-off with a lasting
ceiling (e.g. a `ponytail:` simplification).

### 5.2 Format

Store ADRs in `.ai-memory/adr/` (or `docs/adr/`), numbered sequentially:

```md
# ADR 000N: <Title>

## Status
Accepted | Proposed | Superseded by ADR 000M

## Context
What problem are we solving, and why now?

## Decision
What we chose, in one sentence.

## Consequences
What gets easier, what gets harder, and what ceilings we accept.
```

Template: `.ai-memory/adr/0000-template.md`.

---

## 6. Documentation Discipline

- **Bilingual where required.** Ship docs as `name.md` + `name.zh.md` + an
  `name.i18n.yaml` key catalog so translations stay in sync (harness convention).
- **Safety notices for experimental features.** Mark experimental/unsafe surfaces
  explicitly (mirror the harness `SAFETY.md` bilingual notice).
- **Keep `AGENTS.md` and docs consistent.** If you change agent behavior, update
  this file; if you change architecture, update `docs/` and the relevant ADR.
- **Line-ending and formatting hygiene.** Enforce LF (` .gitattributes`) and final
  newlines (`.editorconfig`) so diffs stay clean across platforms.

---

## 7. Memory Layout

This repository keeps agent-facing knowledge in two tiers:

- **`AGENTS.md`** (this file) — the single source of truth for *how to operate*.
- **`.ai-memory/`** — supporting detail the agent reads on demand:

| Path | Contents |
| --- | --- |
| `.ai-memory/references.md` | Summaries of the three source projects and where their ideas live. |
| `.ai-memory/planning-template.md` | Copy-ready `task_plan.md` / `findings.md` / `progress.md` templates. |
| `.ai-memory/ponytail-ladder.md` | The reuse ladder and `ponytail:` marker rules, in full. |
| `.ai-memory/capability-seams.md` | The Capability Seam pattern and a provider checklist. |
| `.ai-memory/safety-checklist.md` | The non-negotiable safety items, as a review checklist. |
| `.ai-memory/adr/` | Architecture Decision Records (see §5). |

Treat `.ai-memory/` as **durable working memory**: read it at the start of a
task, update it as you learn, and rely on it to resume after a reset (§1.4).

---

## 8. Quick Reference (cheat sheet)

On every turn, in order:

1. **Plan on disk** — 3 files if the task is non-trivial (§1).
2. **Understand first** — read and trace real code before simplifying (§2).
3. **Reuse ladder** — YAGNI → reuse → stdlib → native → dep → one line → minimum (§2.1).
4. **Swap, don't rewrite** — express new behavior as a plugin/provider (§3).
5. **Verify** — run the repo's gates, log the result, never skip (§4).
6. **Record** — ADR for lasting decisions; `ponytail:` for bounded cuts (§5, §2.3).
7. **Resume-safe** — everything important is already in a file.

### Do / Don't

| Do | Don't |
| --- | --- |
| Persist plan + findings + progress to files. | Keep the plan only in chat context. |
| Reuse existing code, stdlib, and platform features. | Rewrite what already works to "make it cleaner." |
| Compose behavior as swappable plugins. | Hardcode a fork into core for one use case. |
| Run the full gate suite before "done." | Declare success on a hunch. |
| Cut code, never safety. | Drop validation, error handling, or a11y to save lines. |
| Log every error and never repeat a failure. | Retry the same broken approach blindly. |
