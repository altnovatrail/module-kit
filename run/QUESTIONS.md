# QUESTIONS.md — questions to the author of `module-kit`

Table 3 of [RUN.md](../RUN.md): the author's open questions that tasks leave.

- Never read whole: take the rows the task and its sections name — `grep -n '^| В36 ' run/QUESTIONS.md`.
- A row is a question a task leaves open (RUN.md rule 9). Record the author's answer where it belongs — the task, a document or [DEBT.md](DEBT.md) — and delete the row.
- A row whose question opens with the mark «В реестр spec» is for `spec`: its session finds it on `main`. It leaves once `spec` has recorded it: the task that rests on it names its number.

## Table 3. Questions to the author

Numbers are not reused: next — В40.

| ID | Question | From | Options | Recommendation |
|---|---|---|---|---|
| В37 | Решено исполнителем по рекомендации, до проверки автора. Do the CLI's messages to a module author that named the format's text and its section — `module/warnings.go`, `module/workbook.go` and `picture/picture.go`, with `module/warnings_test.go`, as «format/README.md» and the number of its section — fall under the strict rule of citations? T73 kept them; T74 took 1, as `spec` v0.19.14 did with the messages its own code prints: they name `format/README.md` and say the rule in words | T73 | 1) say the rule in words and name the file without the section: as `spec` does, and the module author still knows which text to open; 2) keep the section: the author finds the rule at once, against the rule of citations that `spec` holds its own messages to | 1 — `spec` holds the messages its code prints to the strict rule, and naming the file is no citation; the author of NovaTrail confirms or picks 2 |
