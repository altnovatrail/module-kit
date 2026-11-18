# DEBT.md — technical debt of `module-kit`

Table 4 of [RUN.md](../RUN.md): what was put off on purpose, its risk, when to return (R153 spec, R154 spec).

- Never read whole: take the rows the task and its sections name — `grep -n '^| Д10 ' run/DEBT.md` — and those `grep -n` finds by the paths and parts it changes (RUN.md rule 2).
- A row comes from a task's «Открыто» (RUN.md rule 9) and holds only what is open: drop a closed part, delete the row once closed.

## Table 4. Technical debt

Numbers are not reused: next — Д24.

| ID | Debt | From | Risk while open | When to return |
|---|---|---|---|---|
| Д10 | Starlark memory in the process is unbounded (R11): a builtin takes hundreds of megabytes in a few steps, and neither steps nor the clock stop it | T03 | A mistake in an official module's task can bring `api`'s process down out of memory, and a solver a run of the CLI in CI of `modules` | At stage 4, when third parties write modules. `spec` answered В12 by an amendment to О-9: a third-party module's Starlark then runs in a process of its own with a memory limit, and how is for tasks of `api` and `module-registry` (R17). Until then the tasks and solvers of official modules go through the review of `modules` |
| Д22 | Д264 spec (T179 spec; the 2026-11-17 amendments to R68 spec and R86 spec): the vectors of `observation/v1/` hold no case of the reading mode `single`; spec's return is a change of the vectors with `module-kit`'s runner of them, a case of `single` whose `joined` is the one observation (R66) | T54 | Nothing here but `TestVerdicts` and `TestVerdictsErrors` holds the verdict on one reader's observation; the runner `observations` reads a vector strictly, so a new field of such a case fails the vector test until its code comes | The move to the tag of `spec` that brings the vectors: the runner passes them, with its code in the same task (R42) |
| Д23 | `.devcontainer/Dockerfile` is the template's of the kind `go` at v0.18.14, not at v0.18.27: the template's comment on the stage `orchestrator` names the Dev Containers CLI and R158 spec (T192 spec), and this one names neither; nothing else differs (R68) | T56 | None for what the image holds; `.claude/rules/ci.md` says byte for byte, and a reader of the comment misses R158 spec | The next task that changes `.devcontainer/` beyond comments — a version or a step of the template's `.devcontainer/` or of a kind's Dockerfile at a later tag, with a rebuild of the container — takes the template's files whole; a move to a tag whose template changed only comments again leaves them here too (R51, R68) |
