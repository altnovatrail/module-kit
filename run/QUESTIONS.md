# QUESTIONS.md — questions to the author of `module-kit`

Table 3 of [RUN.md](../RUN.md): the author's open questions from tasks' «Открыто» (R153 spec, R154 spec).

- Never read whole: take the rows the task and its sections name — `grep -n '^| В36 ' run/QUESTIONS.md`.
- A row comes from a task's «Открыто» (RUN.md rule 9). Record the author's answer where it belongs — the task, a document or [DEBT.md](DEBT.md) — and delete the row.
- A row whose question opens with the mark «В реестр spec» is for `spec`: its session finds it on `main` (R81 spec; RULES.md, Across repositories). It leaves once `spec` has recorded it: the task that rests on it names its number.

## Table 3. Questions to the author

Numbers are not reused: next — В37.

| ID | Question | From | Options | Recommendation |
|---|---|---|---|---|
| В36 | «В реестр spec» A shared step set of `.github` for a caller's own job in the `toolchain` stage. "Binaries / build" copies six steps of the job `arm64` of the shared `go.yml` — the image and the Go caches — and a guard holds the copy to the pinned `go.yml` (R58, R61): a move of `.github` that changes them fails the required check "Binaries" until a pull request edits the copy by hand, so a grouped pull request of Dependabot cannot pass alone, and a merge by green checks waits for a session. | T48, `/code-review high` | 1) `.github` ships a composite action — the image of `toolchain` and the Go caches from the cache, and their save on `main` — taken by SHA in the group `shared-workflows`; 2) keep the copy and the guard in each caller; 3) an input of `go.yml` that runs a caller's recipe in a job of its own, with its artifact. | 1: the copy and the guard go, Dependabot's pull request carries the steps, and every repository with a job of its own in `toolchain` takes the same steps. |
