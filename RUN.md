# RUN.md — work plan of module-kit

What module-kit does, does not do and has still to decide — [docs/architecture.md](docs/architecture.md); why — [docs/decisions.md](docs/decisions.md). The platform is in `spec/`, a submodule at tag v0.19.12: the component's card — architecture 3.7 spec; the order of repositories — repositories 6 spec. Small tasks, one at a time, gathered in epics; only the current epic is in detail (R19 spec). Other tables, each a file opening with its rules (R153 spec): 2, epics for later — [run/EPICS.md](run/EPICS.md); 3, questions to the author — [run/QUESTIONS.md](run/QUESTIONS.md); 4, debt — [run/DEBT.md](run/DEBT.md). No `run/TASKS.md`: a task outside an epic stands in table 1 (R56, R57).

Language (R154 spec): rules and tables in English, terse, for the session; the epic's description, table 1 and the bodies of tasks in Russian — the author writes and reviews them.

## Rules

A task starts with «Выполни задачу T01 из RUN.md» in a new session in the devcontainer. One task — one session (R143 spec): the hook `UserPromptSubmit` of [.claude/settings.json](.claude/settings.json) — `spec/tools/session/` — stops «Выполни задачу T…» or «планируй T…» of another task in the same session and names `/clear`; the same task passes.

1. Do only the named task; touch other tasks' files only where it requires.
2. Read first: the task, the sections it cites — its own and spec's —, [docs/decisions.md](docs/decisions.md).
   - Conflict with the documents or something missing — stop and ask; never guess.
   - A gap in the component itself — a subsection of [docs/architecture.md](docs/architecture.md), section 2. A gap across repositories — a row of [run/QUESTIONS.md](run/QUESTIONS.md) whose question opens with the mark «В реестр spec» (R81 spec).
   - From [run/DEBT.md](run/DEBT.md) and [run/QUESTIONS.md](run/QUESTIONS.md) take only the rows the task and its sections name and those `grep -n` finds by the paths and parts it changes (R143 spec).
3. Devcontainer only — RULES.md, Hard rules; if `spec/` is empty — CLAUDE.md.
4. Outside this repository — editing `spec` or other repositories, their settings, anything that costs money — only after a warning (what, where, cost) and the author's confirmation. `spec/` is read only here; a task of its own moves it to a new tag. The exception of R146 spec, a task's own checks in a browser or with a real AI call, does not arise: `module-kit` has neither (architecture 1).
5. External facts — from the primary source, with quote and date read (RULES.md); what is not confirmed is an open question.
6. End with the checks of what the task changed; documents — `just -f spec/justfile docs-check-repo`, which holds the rule of citations of `spec/RULES.md` once [docs-budgets.json](docs-budgets.json) sets `"citations": true` (R32 spec). Then `/code-review high`, one round; a small change may skip the review.
7. Do not commit: the author commits after review. A run of spec's orchestrator, «Сделай эпик E…», is the author's word otherwise: its executor works by the contract `spec/tools/orchestrator/executor.md` — commits, pushes and opens the pull request, never merges (R156 spec).
8. End with a short report in Russian: what was done, files, how to check, what the review found or why there was none, what is open and what goes «в реестр spec».
9. When done: table 1 `[ ]` → `[x]`; append to the body «Сделано (дата)» — one line, what was done and its pull request, which says how it was checked; where the author commits after review, the report says it. What stays open is rows, not text: a question for the author, of [run/QUESTIONS.md](run/QUESTIONS.md); what is put off, of [run/DEBT.md](run/DEBT.md) (R160 spec).
   - Then move the body verbatim, relative links fixed, to the epic archive `docs/plan/<epic>.md`, with a gist sentence — an unheaded list item at the start of the receiving section: «T01 — сделана <дата>: <итог одной фразой>». Neither stays here, and a search by heading finds one body, in the archive (R153 spec).
10. Language (RULES.md spec, Hard rules; R154 spec): this plan and `run/` in English, the epic's description, table 1, the bodies of tasks, the epic archives `docs/plan/` and the reports in Russian; `docs/`, README, CLAUDE.md, code and comments in English; commit messages and a pull request's title and description in English and terse, for the session, never a copy of the report.
11. A task larger than one review (≈ 30–60 minutes of reading): propose a split before starting.
12. Before a new task, close the findings on the previous one (R30 spec); an allowed departure from rule 1.
    - SonarCloud — [New Code](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit), after the merge; CodeQL and Scorecard — `gh api 'repos/altnovatrail/module-kit/code-scanning/alerts?state=open'`.
    - Fix in code first; "false positive" is the last resort, with the reason in words (R33 spec or a decision of this log, like R51 canvas-api). Never skip silently.
    - Report these fixes in a separate paragraph: the author commits them apart from the task.

**Statuses:** `[ ]` not started; `[x]` done; `[-]` not done, reason in the task; `[?]` deferred, reason and when to return in the task. A task that fails the author's review returns to `[ ]` with what to fix.

**Who:** «Claude» — the executor does it, the author reviews; «Claude, проверяю я» — texts the author reads line by line; «я + Claude» — the author does some steps: decisions, sign-ins, GitHub settings, confirmations by rule 4.

**Plan:** «да» — plan mode first: decisions missing from the documents, or larger than one review (rule 11); «—» — run at once.

**Task body** (Russian): «Цель», «Сделать», «Как проверить», «Разделы»; as needed «Не делать», «Решить при планировании», «Мой шаг», «Стоимость», «Риски»; from planning «Почему», «Решено автором (дата)», «Выяснилось при планировании»; when done «Сделано (дата)» (rule 9).

## Epics

- **Table 1** is the current epic: its tasks in detail, run one at a time.
- **No tasks outside an epic** (R56): a task that does not come from an epic — a move to a tag of `spec`, an analysis by R145 spec — stands in table 1 of the current epic; with none, it is the first task of the next, and a task the author starts between epics starts the next epic, as T45 started E3 (R57).
- **Next epic.** Once table 1 is all checked, the author picks the next epic from [run/EPICS.md](run/EPICS.md) and splits it with Claude in plan mode into tasks of one review each, into table 1 and this file.
- **A done epic** goes to `docs/plan/<epic>.md` (R59 canvas-api): bodies and gist sentences are there already (rule 9); move its description and table of tasks verbatim; questions and debt stay in `run/`; until the next epic, table 1 reads «Current epic — none» and names the last done epic. Measure the epic's sessions — `just -f spec/justfile tokens-report --since <first day of the epic> --row` — as a row of [docs/measurements.md](docs/measurements.md), after measurements 1.9 spec; a new analysis by the threshold of R145 spec.
- Never read an archive whole: find a task by heading — `grep -n '^#### T03\.' RUN.md docs/plan/*.md`; from another repository — «T03 module-kit». Task numbers are global, never reused, and taken free at the moment of writing; a task's number and a dated line stand only here, in `run/` and in `docs/plan/`; outside them a text cites no decision and no section, and only a claim about the outside world cites its record — `sources 1.1`, spec's `sources 2.3 spec` (R171 spec).

## Table 1. Current epic — E13. Переходы на `spec` v0.19.11 и v0.19.12, ссылки по R171 spec

| Статус | ID | Задача | Кто | План | Зависит от |
|---|---|---|---|---|---|
| [x] | T64 | Переход на тег `spec` v0.19.11 — тело в [docs/plan/E13.md](docs/plan/E13.md) | Claude | — | — |
| [x] | T65 | Ссылки по R171 spec, часть 1: план, CLAUDE.md, `.claude/rules/`, комментарии кода — тело в [docs/plan/E13.md](docs/plan/E13.md) | Claude | да | T64 |
| [x] | T72 | Переход на тег `spec` v0.19.12 — тело в [docs/plan/E13.md](docs/plan/E13.md) | Claude | — | T65 |
| [x] | T73 | Ссылки по строгому правилу R171 spec: проход по первой части — тело в [docs/plan/E13.md](docs/plan/E13.md) | Claude | — | T72 |
| [x] | T66 | Ссылки по R171 spec, часть 2: `docs/` без журнала решений — тело в [docs/plan/E13.md](docs/plan/E13.md) | Claude | — | T65 |
| [x] | T67 | Ссылки по R171 spec, часть 3: журнал решений — вводная, R01–R08 — тело в [docs/plan/E13.md](docs/plan/E13.md) | Claude | — | T65 |
| [x] | T68 | Ссылки по R171 spec, часть 4: журнал решений — R09–R20 — тело в [docs/plan/E13.md](docs/plan/E13.md) | Claude | — | T65 |
| [ ] | T69 | Ссылки по R171 spec, часть 5: журнал решений — R21–R29 | Claude | — | T65 |
| [ ] | T70 | Ссылки по R171 spec, часть 6: журнал решений — R30–R43 | Claude | — | T65 |
| [ ] | T71 | Ссылки по R171 spec, часть 7: журнал решений — R44–R69, ключ `"citations": true` | Claude | — | T66–T70 |

## Epic E13. Переходы на `spec` v0.19.11 и v0.19.12, ссылки по R171 spec

**Почему эпик такой.** Прогон «Сделай эпик E22» оркестратором `spec` (R156 spec): его волна 0 переводит репозиторий на последний тег `spec`, v0.19.11, затем строка 7.39 COORDINATION.md spec — ссылки по R171 spec, по тексту T259 spec. Задачи — пачка на ветке `e22-module-kit` с одним pull request (R159 spec); их тексты дал оркестратор (правило 6 COORDINATION.md spec). Текущего эпика не было, а задача не из эпика — первая задача следующего (R56): T64 начала эпик E13, как T63 начала E12 (R57). Второй прогон эпика E22 — пачка на ветке `e22-module-kit-2`: его волна 0, T72, переводит репозиторий на v0.19.12, где R171 spec строже — вне плана и архивов ни номера решения, ни раздела, ни документа, только запись `sources` или `measurements` о внешнем мире, — и части T66–T71 идут по R171 spec в этой редакции; T73 проходит по ней то, что правила первая часть, T65.

**Не входит:** новый код библиотеки и CLI — формат тот же, 0.4.

Части T66–T71 — задача 7.39 COORDINATION.md spec, разбитая по шагу 9 её текста: он и общие шаги — в теле T65, [docs/plan/E13.md](docs/plan/E13.md). Каждая часть берёт механические места командой `docs-check-repo` с `"citations": true` во временной копии `docs-budgets.json`, которую не коммитит, — до T71, которая включает ключ, — а остальное — `grep -nE` по шаблону measurements 1.18 spec; каждую ссылку — по шагам 4–8 тела T65; в отчёте — ссылок было и стало, по видам, командой measurements 1.18 spec. Число мест в скобках — по проверке с ключом на v0.19.11; на v0.19.12 та же проверка находит 3 259 мест: `docs/decisions.md` — 2 336, `docs/architecture.md` — 643, `docs/sources.md` — 107, `docs/measurements.md` — 60, `docs/README.md` — 8, CLAUDE.md — 25, `.claude/rules/` — 26, остальное — 54 в коде, justfile, `.github/` и копиях файлов формата.

#### T69. Ссылки по R171 spec, часть 5: журнал решений — R21–R29

- **Цель:** записи R21–R29, разделы `##` от «The YAML reader» до «The observation», — по R171 spec, их датированные поправки свёрнуты в действующий текст.
- **Сделать:** механические места (107) и остальные ссылки, свёртка поправок и противоречия — по шагам 2–3 T67.
- **Как проверить:** как у T67.
- **Разделы:** R171 spec, R160 spec; тело T65.

#### T70. Ссылки по R171 spec, часть 6: журнал решений — R30–R43

- **Цель:** записи R30–R43, разделы `##` от «The leak check» до «The command `open`», — по R171 spec, их датированные поправки свёрнуты в действующий текст.
- **Сделать:** механические места (149) и остальные ссылки, свёртка поправок и противоречия — по шагам 2–3 T67.
- **Как проверить:** как у T67.
- **Разделы:** R171 spec, R160 spec; тело T65.

#### T71. Ссылки по R171 spec, часть 7: журнал решений — R44–R69, ключ `"citations": true`

- **Цель:** записи R44–R69, разделы `##` от «The devcontainer image on arm64» до конца журнала, — по R171 spec, их датированные поправки свёрнуты в действующий текст; `docs-budgets.json` — с `"citations": true`, как `spec/repo-template/files/docs-budgets.json`, и `docs-check-repo` с ним — зелёный.
- **Сделать:**
  1. Механические места (239) и остальные ссылки, свёртка поправок и противоречия — по шагам 2–3 T67.
  2. `docs-budgets.json` — ключ `"citations": true`; места, которые проверка с ним найдёт после T66–T70, — по шагам 4–8 тела T65.
  3. Строка Д24 закрыта, когда в журнале не осталось датированных поправок: удалить её.
- **Как проверить:** `docs-check-repo` с ключом — зелёный; `/code-review high`, один круг; в описании pull request — таблица «поправка → место в новом тексте»; в отчёте — ссылок было до T65 и стало после T71, по видам, командой measurements 1.18 spec.
- **Разделы:** R171 spec, R160 spec; `spec/repo-template/files/docs-budgets.json`; тело T65; Д24.
