# DEBT.md — technical debt of `module-kit`

Table 4 of [RUN.md](../RUN.md): what was put off on purpose, its risk, when to return (R153 spec, R154 spec).

- Never read whole: take the rows the task and its sections name — `grep -n '^| Д10 ' run/DEBT.md` — and those `grep -n` finds by the paths and parts it changes (RUN.md rule 2).
- A row comes from a task's «Открыто» (RUN.md rule 9) and holds only what is open: drop a closed part, delete the row once closed.

## Table 4. Technical debt

Numbers are not reused: next — Д22.

| ID | Debt | From | Risk while open | When to return |
|---|---|---|---|---|
| Д10 | Starlark memory in the process is unbounded (R11): a builtin takes hundreds of megabytes in a few steps, and neither steps nor the clock stop it | T03 | A mistake in an official module's task can bring `api`'s process down out of memory, and a solver a run of the CLI in CI of `modules` | At stage 4, when third parties write modules. `spec` answered В12 by an amendment to О-9: a third-party module's Starlark then runs in a process of its own with a memory limit, and how is for tasks of `api` and `module-registry` (R17). Until then the tasks and solvers of official modules go through the review of `modules` |
