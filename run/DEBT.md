# DEBT.md — technical debt of `module-kit`

Table 4 of [RUN.md](../RUN.md): what was put off on purpose, its risk, when to return.

- Never read whole: take the rows the task and its sections name — `grep -n '^| Д10 ' run/DEBT.md` — and those `grep -n` finds by the paths and parts it changes (RUN.md rule 2).
- A row is what a task leaves open (RUN.md rule 9) and holds only what is open: drop a closed part, delete the row once closed.

## Table 4. Technical debt

Numbers are not reused: next — Д26.

| ID | Debt | From | Risk while open | When to return |
|---|---|---|---|---|
| Д10 | Starlark memory in the process is unbounded (R11): a builtin takes hundreds of megabytes in a few steps, and neither steps nor the clock stop it | T03 | A mistake in an official module's task can bring `api`'s process down out of memory, and a solver a run of the CLI in CI of `modules` | At stage 4, when third parties write modules. `spec` answered В12 in О-9: a third-party module's Starlark then runs in a process of its own with a memory limit, and how is for tasks of `api` and `module-registry`. Until then the tasks and solvers of official modules go through the review of `modules` |
| Д22 | Д264 spec: the vectors of `observation/v1/` hold no case of the reading mode `single`; spec's return is a change of the vectors with `module-kit`'s runner of them, a case of `single` whose `joined` is the one observation (R66) | T54 | Nothing here but `TestVerdicts` and `TestVerdictsErrors` holds the verdict on one reader's observation; the runner `observations` reads a vector strictly, so a new field of such a case fails the vector test until its code comes | The move to the tag of `spec` that brings the vectors: the runner passes them, with its code in the same task (R42) |
| Д25 | `docs-budgets.json` has no `"citations": true`: with it `docs-check-repo` finds 8 places, all in the byte-for-byte copies of spec's format files — the descriptions of `internal/schema/schemas/canonical-view.schema.json`, `module-index.schema.json` and `template-vector.schema.json`, and the comment of `page/page.yaml` —, which cite sections of the format's text and a decision of `canvas`; the rest of the repository passes the check with the key (В38, В39) | T71 | CI does not hold the rule of citations here: a new citation in a document or a comment passes the documents check until the key is on; a check with the key in a temporary copy of `docs-budgets.json` finds it | The move to the tag of `spec` whose `format/schemas/` and `format/page.yaml` cite no section: the copies come with it, and the same task turns the key on |
