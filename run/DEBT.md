# DEBT.md — technical debt of `module-kit`

Table 4 of [RUN.md](../RUN.md): what was put off on purpose, its risk, when to return (R153 spec, R154 spec).

- Never read whole: take the rows the task and its sections name — `grep -n '^| Д10 ' run/DEBT.md` — and those `grep -n` finds by the paths and parts it changes (RUN.md rule 2).
- A row is what a task leaves open (RUN.md rule 9) and holds only what is open: drop a closed part, delete the row once closed.

## Table 4. Technical debt

Numbers are not reused: next — Д25.

| ID | Debt | From | Risk while open | When to return |
|---|---|---|---|---|
| Д10 | Starlark memory in the process is unbounded (R11): a builtin takes hundreds of megabytes in a few steps, and neither steps nor the clock stop it | T03 | A mistake in an official module's task can bring `api`'s process down out of memory, and a solver a run of the CLI in CI of `modules` | At stage 4, when third parties write modules. `spec` answered В12 by an amendment to О-9: a third-party module's Starlark then runs in a process of its own with a memory limit, and how is for tasks of `api` and `module-registry` (R17). Until then the tasks and solvers of official modules go through the review of `modules` |
| Д22 | Д264 spec (T179 spec; the 2026-11-17 amendments to R68 spec and R86 spec): the vectors of `observation/v1/` hold no case of the reading mode `single`; spec's return is a change of the vectors with `module-kit`'s runner of them, a case of `single` whose `joined` is the one observation (R66) | T54 | Nothing here but `TestVerdicts` and `TestVerdictsErrors` holds the verdict on one reader's observation; the runner `observations` reads a vector strictly, so a new field of such a case fails the vector test until its code comes | The move to the tag of `spec` that brings the vectors: the runner passes them, with its code in the same task (R42) |
| Д24 | Dated amendments name an earlier tag of `spec` as the current one, «`module-kit` is on `spec` v0.18.27» — in R01–R03, R05, R08, R12, R14, R53, R54, R57, R59, R62, R63 and R66 —, and those of R51, R54, R57 and R63 name the Dockerfile's comment of a row Д23 that T63 took with v0.19.4 and closed. A move to a tag is its commit, not an entry (R160 spec), and the entries of moves — R47–R49, R51–R54, R57, R62, R63, R66 — keep of themselves what is still a decision in force: the template's `.devcontainer/` (R51, R54), the plan's layout (R57), the stages of the Dockerfile (R62), `shutdownAction` (R63) | T63 | A session that reads one of them takes a past tag, or the comment of Д23, for the current state; the tag is named by CLAUDE.md, RUN.md and `go.mod`, the devcontainer by R68, architecture 3.1 and `.claude/rules/ci.md` | The task that touches one of them rewrites it whole (R160 spec); all of them at once is a task of its own, larger than one review, by the author's word or at the next analysis by R145 spec |
