# DEBT.md — technical debt of `module-kit`

Table 4 of [RUN.md](../RUN.md): what was put off on purpose, its risk, when to return (R153 spec, R154 spec).

- Never read whole: take the rows the task and its sections name — `grep -n '^| Д10 ' run/DEBT.md` — and those `grep -n` finds by the paths and parts it changes (RUN.md rule 2).
- A row comes from a task's «Открыто» (RUN.md rule 9) and holds only what is open: drop a closed part, delete the row once closed.

## Table 4. Technical debt

Numbers are not reused: next — Д22.

| ID | Debt | From | Risk while open | When to return |
|---|---|---|---|---|
| Д10 | Starlark memory in the process is unbounded (R11): a builtin takes hundreds of megabytes in a few steps, and neither steps nor the clock stop it | T03 | A mistake in an official module's task can bring `api`'s process down out of memory, and a solver a run of the CLI in CI of `modules` | At stage 4, when third parties write modules. `spec` answered В12 by an amendment to О-9: a third-party module's Starlark then runs in a process of its own with a memory limit, and how is for tasks of `api` and `module-registry` (R17). Until then the tasks and solvers of official modules go through the review of `modules` |
| Д21 | "Binaries / build" keeps its copy of the six steps «Toolchain image…» and «Go caches…» of the job `arm64` of the shared `go.yml` and the guard «Cache steps as in the shared go.yml» (R58, R61). `spec` recorded В36 by its option 1 as rows 2.5 and 3.11 of `COORDINATION.md` spec: once `.github` ships the composite action of row 2.5, the job takes it by SHA, the copy and the guard go, and an amendment to R58 records it (the 2026-11-16 amendment to R61; R63) | T51; В36 | A pull request of Dependabot that moves `.github` and changes those steps fails "Binaries" until a session edits the copy, so a merge by green checks waits | The task of row 3.11 of `COORDINATION.md` spec, after the release of `.github` with row 2.5 |
