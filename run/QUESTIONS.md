# QUESTIONS.md — questions to the author of `module-kit`

Table 3 of [RUN.md](../RUN.md): the author's open questions that tasks leave.

- Never read whole: take the rows the task and its sections name — `grep -n '^| В36 ' run/QUESTIONS.md`.
- A row is a question a task leaves open (RUN.md rule 9). Record the author's answer where it belongs — the task, a document or [DEBT.md](DEBT.md) — and delete the row.
- A row whose question opens with the mark «В реестр spec» is for `spec`: its session finds it on `main`. It leaves once `spec` has recorded it: the task that rests on it names its number.

## Table 3. Questions to the author

Numbers are not reused: next — В40.

| ID | Question | From | Options | Recommendation |
|---|---|---|---|---|
| В37 | Решено исполнителем по рекомендации, до проверки автора. Do the CLI's messages to a module author that name the format's text and its section — `module/warnings.go`, `module/workbook.go` and `picture/picture.go`, with `module/warnings_test.go`, as «format/README.md» and the number of its section — fall under the strict rule of citations? T73 kept them. | T73 | 1) keep: they are what the CLI prints to a module author, pointing to the normative text where the rule stands, which the author opens to act; changing them changes the CLI's output and its tests; 2) say the rule in words without the section: the author loses the pointer to the text | 1 — the rule is about the citations of documents and code comments, and these are the CLI's output to its user, the module author; the author of NovaTrail confirms or picks 2 |
| В38 | В реестр spec. `internal/schema/schemas/*.json` and `page/page.yaml` are byte-for-byte copies of spec's `format/schemas/` and `format/page.yaml`, which tests hold to spec's; their descriptions and comment cite sections of the format's text (8 places by the check with `"citations": true`), which this repository cannot change. Who removes them before T71 turns the key on? | T73 | 1) spec rewrites the descriptions of `format/schemas/` and the comment of `format/page.yaml` without sections, and `module-kit` takes them with its next tag of spec; 2) `docs-check-repo` skips a file byte-identical to a file of spec at the submodule's tag; 3) `module-kit` keeps the key off until spec's tag brings them | 1 — the files are spec's, the rule covers schema descriptions, and every repository that copies them gets them fixed at once |
| В39 | Решено исполнителем по рекомендации, до проверки автора. T71 was to turn `"citations": true` on in `docs-budgets.json` with `docs-check-repo` green, but with the key the check fails on 8 places in the copies of spec's files that В38 names, which tests hold byte for byte to `spec.Format` and this repository cannot change; "Documents / docs-check-repo" is a required check. What does T71 do with the key? | T71 | 1) leave the key off; every other place of the repository passes the check with the key, and Д25 turns it on with the move to the tag of `spec` that brings the copies without sections; 2) turn the key on with the check red: the pull request cannot merge until `spec` changes; 3) edit the copies and let their test skip descriptions and comments: the copies would no longer be spec's, against what the test holds | 1 — the rest of the rule holds now, the copies are spec's to fix (В38), and a red required check would stop the epic's merge |
