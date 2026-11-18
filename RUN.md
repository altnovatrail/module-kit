# RUN.md — work plan of module-kit

What module-kit does, does not do and has still to decide — [docs/architecture.md](docs/architecture.md); why — [docs/decisions.md](docs/decisions.md). The platform is in `spec/`, a submodule at tag v0.18.34: the component's card — architecture 3.7 spec; the order of repositories — repositories 6 spec. Small tasks, one at a time, gathered in epics; only the current epic is in detail (R19 spec). Other tables, each a file opening with its rules (R153 spec): 2, epics for later — [run/EPICS.md](run/EPICS.md); 3, questions to the author — [run/QUESTIONS.md](run/QUESTIONS.md); 4, debt — [run/DEBT.md](run/DEBT.md). No `run/TASKS.md`: a task outside an epic stands in table 1 (R56, R57).

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
6. End with the checks of what the task changed; documents — `just -f spec/justfile docs-check-repo` (R32 spec). Then `/code-review high`, one round; a small change may skip the review.
7. Do not commit: the author commits after review. A run of spec's orchestrator, «Сделай эпик E…», is the author's word otherwise: its executor works by the contract `spec/tools/orchestrator/executor.md` — commits, pushes and opens the pull request, never merges (R156 spec).
8. End with a short report in Russian: what was done, files, how to check, what the review found or why there was none, what is open and what goes «в реестр spec».
9. When done: table 1 `[ ]` → `[x]`; append to the body «Сделано (дата)» — one line, what was done and its pull request, which says how it was checked; where the author commits after review, the report says it. What stays open is rows, not text: a question for the author, of [run/QUESTIONS.md](run/QUESTIONS.md); what is put off, of [run/DEBT.md](run/DEBT.md) (R160 spec).
   - Then move the body verbatim, relative links fixed, to the epic archive `docs/plan/<epic>.md`, with a gist sentence — an unheaded list item at the start of the receiving section: «T01 — сделана <дата>: <итог одной фразой>». Neither stays here, and a search by heading finds one body, in the archive (R153 spec).
10. Language (R143 spec, R154 spec): this plan and `run/` in English, the epic's description, table 1, the bodies of tasks, the epic archives `docs/plan/` and the reports in Russian; `docs/`, README, CLAUDE.md, code and comments in English; commit messages and a pull request's title and description in English and terse, for the session, never a copy of the report (the 2026-11-17 amendment to R14 spec).
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
- Never read an archive whole: find a task by heading — `grep -n '^#### T03\.' RUN.md docs/plan/*.md`; from another repository — «T03 module-kit». Task numbers are global, never reused, and taken free at the moment of writing.

## Table 1. Current epic — E11. Тексты без истории (R160 spec)

| Статус | ID | Задача | Кто | План | Зависит от |
|---|---|---|---|---|---|
| [x] | T57 | Тексты без истории (R160 spec) и тег `spec` v0.18.34 — тело в [docs/plan/E11.md](docs/plan/E11.md) | Claude | — | — |
| [x] | T58 | Документы без истории: `docs/architecture.md` без раздела 3.6, `docs/README.md`, README, CLAUDE.md, `.claude/rules/` — тело в [docs/plan/E11.md](docs/plan/E11.md) | Claude | — | T57 |
| [ ] | T59 | Документы без истории: раздел 3.6 `docs/architecture.md` | Claude | — | T57 |
| [ ] | T60 | Комментарии кода по R160 spec: `module`, `template`, `cmd/module-kit` | Claude | — | T57 |
| [ ] | T61 | Комментарии кода по R160 spec: `moduleversion`, `moduleindex`, `modulepackage`, `observation`, `verdict`, `leak`, `openview`, `page`, `picture` | Claude | — | T57 |
| [ ] | T62 | Комментарии кода по R160 spec: `internal/`, `.github/`, `justfile` | Claude | — | T57 |

## Epic E11. Тексты без истории (R160 spec)

**Почему эпик такой.** Строка 7.27 COORDINATION.md spec — правило эпика E16 `spec` в этом репозитории, по тексту T206 spec: тег `spec` с T205, `docs-budgets.json` с `entry` 12 и `comments`, план, документы и комментарии кода без истории (R160 spec; R76 spec, R156 spec). Текст задачи дал оркестратор (правило 6 COORDINATION.md spec). Текущего эпика не было, а задача не из эпика — первая задача следующего (R56): T57 начала эпик E11, как T56 начала E10 (R57). Всё вместе — больше одного ревью: 149 КБ `docs/architecture.md` и около 300 КБ комментариев Go. T57 разделила его по правилу 11: сама — тег, `docs-budgets.json` и план; документы — T58 и T59, комментарии кода — T60, T61 и T62, по пакетам, примерно по 100 КБ комментариев. Задачи вне эпика стоят в таблице 1, а не в `run/TASKS.md`, которого здесь нет (R56); оркестратор читает их оттуда (README оркестратора spec, раздел 2).

**Не входит:** новый код библиотеки и CLI — формат тот же; записи журнала целиком — каждую переписывает задача, которая её касается (R160 spec, «Decisions»); `docs/sources.md`, `docs/measurements.md`, архивы `docs/plan/` — вне правила (R160 spec, «Outside the rule»); `.devcontainer/` — это файлы заготовки `spec` (R51), их комментарии приходят переходом на тег.

## Задачи эпика

#### T59. Документы без истории: раздел 3.6 `docs/architecture.md`

- **Цель:** раздел 3.6 `docs/architecture.md` — библиотека и CLI — описывает текущее устройство, без истории (R160 spec, «One home», «A done task»).
- **Почему:** строка 7.27 COORDINATION.md spec, пункт «Документы» текста T206 spec; раздел — 60 КБ, одно ревью сам по себе; T57 выделила его по правилу 11 (R76 spec, R156 spec).
- **Сделать:**
  1. Раздел 3.6 и его подразделы — как пункт 1 T58: без «since T…», «until…», дат и пересказа решений; значение — файлом, где оно задано; решение — номером.
  2. Свой предел `"architecture 3.6": 66` в [docs-budgets.json](docs-budgets.json) — по новому размеру с запасом 10–20% (measurements 1.9 spec) или убрать, если раздел уложился в предел блока `repository`.
  3. Запись журнала, которую правка меняет, — как пункт 4 T58.
- **Не делать:** остальные разделы — T58; комментарии кода — T60–T62.
- **Как проверить:** `just -f spec/justfile docs-check-repo`; `/code-review high`. Больше одного ревью — разделить по правилу 11.
- **Разделы:** R160 spec; RULES.md spec — The bar; R145 spec; текст T206 spec.

#### T60. Комментарии кода по R160 spec: `module`, `template`, `cmd/module-kit`

- **Цель:** комментарии Go пакетов `module`, `template` и `cmd/module-kit`, тесты тоже, — по R160 spec, «Comments»: почему, ограничение, ловушка, контракт экспортируемого API; без пересказа кода, истории, номеров задач и разделов документов; где причина — решение, ограничение одной фразой и номер в конце, `(R25)`.
- **Почему:** строка 7.27 COORDINATION.md spec, пункт «Комментарии» текста T206 spec; около 100 КБ комментариев, одно ревью; T57 выделила его по правилу 11 (R76 spec, R156 spec).
- **Сделать:**
  1. Каждый блок комментария: пересказ кода — убрать; `since T…`, `until v…`, `T37`, `architecture 3.6`, `format/README.md, section 20` — убрать или заменить номером решения; `R… in docs/decisions.md` — `R…`; решение `spec` — `R… spec`; номер, который не держит то, что говорит комментарий, — поправить.
  2. Doc comment пакета и экспортируемого имени — его контракт, для `go doc`; код не меняется.
- **Как проверить:** правило [go](.claude/rules/go.md) — `just ci-lint`, `just ci-test`; `just -f spec/justfile docs-check-repo` — номера в комментариях; `/code-review high`. Больше одного ревью — разделить по правилу 11.
- **Разделы:** R160 spec, R32 spec; RULES.md spec — The bar.

#### T61. Комментарии кода по R160 spec: `moduleversion`, `moduleindex`, `modulepackage`, `observation`, `verdict`, `leak`, `openview`, `page`, `picture`

- **Цель:** комментарии Go этих пакетов, тесты тоже, — как в T60.
- **Почему:** как T60; около 110 КБ комментариев.
- **Сделать:** как пункты 1 и 2 T60, для этих пакетов.
- **Как проверить:** как T60.
- **Разделы:** R160 spec, R32 spec; RULES.md spec — The bar.

#### T62. Комментарии кода по R160 spec: `internal/`, `.github/`, `justfile`

- **Цель:** комментарии Go пакетов `internal/`, тесты тоже, и комментарии `#` `.github/` и `justfile` — как в T60.
- **Почему:** как T60; около 100 КБ комментариев Go и 25 КБ CI.
- **Сделать:** как пункты 1 и 2 T60; в `.github/` и `justfile` — `Д… in RUN.md` строки, которых нет в [run/DEBT.md](run/DEBT.md), — убрать, причину оставить решением.
- **Не делать:** `.devcontainer/` — файлы заготовки `spec` (R51).
- **Как проверить:** правила [go](.claude/rules/go.md) и [ci](.claude/rules/ci.md) — `just ci-lint`, `just ci-test`; `just -f spec/justfile docs-check-repo`; `/code-review high`. Больше одного ревью — разделить по правилу 11.
- **Разделы:** R160 spec, R32 spec; RULES.md spec — The bar.
