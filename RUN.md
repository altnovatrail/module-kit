# RUN.md — work plan of module-kit

What module-kit does, does not do and has still to decide — [docs/architecture.md](docs/architecture.md); why — [docs/decisions.md](docs/decisions.md). The platform is in `spec/`, a submodule at tag v0.18.4: the component's card — architecture 3.7 spec; the order of repositories — repositories 6 spec. Small tasks, one at a time, gathered in epics; only the current epic is in detail (R19 spec). Other tables, each a file opening with its rules (R153 spec): 2, epics for later — [run/EPICS.md](run/EPICS.md); 3, questions to the author — [run/QUESTIONS.md](run/QUESTIONS.md); 4, debt — [run/DEBT.md](run/DEBT.md). No `run/TASKS.md`: a task outside an epic stands in table 1 (R56, R57).

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
9. When done: table 1 `[ ]` → `[x]`; append to the body «Сделано (дата)» — what was done, how checked — and «Открыто». A question for the author from «Открыто» is also a row of [run/QUESTIONS.md](run/QUESTIONS.md); what is put off, a row of [run/DEBT.md](run/DEBT.md).
   - Then move the body verbatim, relative links fixed, to the epic archive `docs/plan/<epic>.md`, with a gist sentence — an unheaded list item at the start of the receiving section: «T01 — сделана <дата>: <итог одной фразой>». Neither stays here, and a search by heading finds one body, in the archive (R153 spec).
10. Language (R143 spec, R154 spec): this plan and `run/` in English, the epic's description, table 1, the bodies of tasks, the epic archives `docs/plan/` and the reports in Russian; `docs/`, README, CLAUDE.md, code and comments in English.
11. A task larger than one review (≈ 30–60 minutes of reading): propose a split before starting.
12. Before a new task, close the findings on the previous one (R30 spec); an allowed departure from rule 1.
    - SonarCloud — [New Code](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit), after the merge; CodeQL and Scorecard — `gh api 'repos/altnovatrail/module-kit/code-scanning/alerts?state=open'`.
    - Fix in code first; "false positive" is the last resort, with the reason in words (R33 spec or a decision of this log, like R51 canvas-api). Never skip silently.
    - Report these fixes in a separate paragraph: the author commits them apart from the task.

**Statuses:** `[ ]` not started; `[x]` done; `[-]` not done, reason in the task; `[?]` deferred, reason and when to return in the task. A task that fails the author's review returns to `[ ]` with what to fix.

**Who:** «Claude» — the executor does it, the author reviews; «Claude, проверяю я» — texts the author reads line by line; «я + Claude» — the author does some steps: decisions, sign-ins, GitHub settings, confirmations by rule 4.

**Plan:** «да» — plan mode first: decisions missing from the documents, or larger than one review (rule 11); «—» — run at once.

**Task body** (Russian): «Цель», «Сделать», «Как проверить», «Разделы»; as needed «Не делать», «Решить при планировании», «Мой шаг», «Стоимость», «Риски»; from planning «Почему», «Решено автором (дата)», «Выяснилось при планировании»; when done «Сделано (дата)», «Открыто» (rule 9).

## Epics

- **Table 1** is the current epic: its tasks in detail, run one at a time.
- **No tasks outside an epic** (R56): a task that does not come from an epic — a move to a tag of `spec`, an analysis by R145 spec — stands in table 1 of the current epic; with none, it is the first task of the next, and a task the author starts between epics starts the next epic, as T45 started E3 (R57).
- **Next epic.** Once table 1 is all checked, the author picks the next epic from [run/EPICS.md](run/EPICS.md) and splits it with Claude in plan mode into tasks of one review each, into table 1 and this file.
- **A done epic** goes to `docs/plan/<epic>.md` (R59 canvas-api): bodies and gist sentences are there already (rule 9); move its description and table of tasks verbatim; questions and debt stay in `run/`; until the next epic, table 1 reads «Current epic — none» and names the last done epic. Measure the epic's sessions — `just -f spec/justfile tokens-report --since <first day of the epic> --row` — as a row of [docs/measurements.md](docs/measurements.md), after measurements 1.9 spec; a new analysis by the threshold of R145 spec.
- Never read an archive whole: find a task by heading — `grep -n '^#### T03\.' RUN.md docs/plan/*.md`; from another repository — «T03 module-kit». Task numbers are global, never reused, and taken free at the moment of writing.

## Table 1. Current epic — E4. Скорость CI: строка 7.18

| Статус | ID | Задача | Кто | План | Зависит от |
|---|---|---|---|---|---|
| [x] | T47 | Скорость CI: строка 7.18 — тело в [docs/plan/E4.md](docs/plan/E4.md) | Claude | — | — |
| [x] | T50 | Д20: тесты на часах предиката — после остальных в `ci-test` — тело в [docs/plan/E4.md](docs/plan/E4.md) | Claude | — | — |
| [x] | T48 | CI — на релизе `.github` строки 7.13 — тело в [docs/plan/E4.md](docs/plan/E4.md) | Claude | — | T47; релиз `.github` строки 7.13 |
| [ ] | T49 | Переход на тег `spec` строки 7.14 — стадия `toolchain-core` | Claude | — | T48; тег `spec` строки 7.14; релиз `.github` с T66 `.github` |

## Epic E4. Скорость CI: строка 7.18

**Почему эпик такой.** Строка 7.18 COORDINATION.md spec — скорость CI `module-kit` по замерам measurements 3.1 spec; слово автора 2026-11-16 — ускорение CI сейчас, в прогоне оркестратора T171 spec (R76 spec, шестая поправка 2026-11-15; R156 spec). Текущего эпика не было, а задача не из эпика — первая задача следующего (R56): T47 начала эпик E4 и сделала то, что не ждёт ни релиза `.github`, ни тега `spec`; T48 и T49 ждут их, и оркестратор запустит каждую, когда выйдет то, чего она ждёт. T49 — последняя: она меняет `.devcontainer/` (R59). T50 — не из строки: ночная работа по слову автора 2026-11-16 закрыла Д20, долг T47, — задача вне эпика стоит в таблице текущего (R56, R60).

**Не входит:** что проверки доказывают — те же случаи, `-race`, покрытие и байты бинарников; код библиотеки и CLI.

#### T49. Переход на тег `spec` строки 7.14 — стадия `toolchain-core`

- **Цель:** `module-kit` на теге `spec` строки 7.14 (T166 spec): `.devcontainer/Dockerfile` со стадией `toolchain-core`, и проверки без браузера и Go-тестов — lint, vuln, licenses, документы — и релиз идут в ней, как их запускает релиз `.github` строки 7.13.
- **Почему:** строка 7.13 COORDINATION.md spec — «с тегом `spec` строки 7.14 — релиз и проверки без браузера и Go-тестов (lint, vuln, licenses, документы) в стадии `toolchain-core`»; переход меняет `.devcontainer/`, и его делает последняя задача эпика (текст T47, шаг 2; R59).
- **Зависит от:** T48 и тега `spec` строки 7.14; и релиза `.github` с T66 `.github` — стадия `toolchain-core` в релизе и проверках, последняя часть строки 7.13: T48 взяла v0.1.30 без неё, потому что T66 нужен переход `.github` на тег `spec` строки 7.14, а он меняет `.devcontainer/`, которую ночью не меняют (T171 spec; R61).
- **Сделать:**
  1. Как T45: `gh release verify <тег> -R altnovatrail/spec`; подмодуль `spec/` и модули Go `spec`, `spec/contracts` — на тег; тег — везде, где репозиторий его называет; что теги после v0.18.3 несут этому репозиторию — `git -C spec log --oneline v0.18.3..<тег>`, `grep -n '`module-kit`' spec/run/DEBT.md`, карточка architecture 3.7 spec.
  2. `.devcontainer/` — байт в байт заготовка `just new-repo-draft module-kit go 3.7 <тег>` на клоне `spec` во временной папке (R51, R54).
  3. Релиз `.github` с T66 — его `uses:` и копия шагов «Binaries / build» из job `arm64` его `go.yml`, как в T48 (R61): если T66 меняет путь или имя файла образа, шаг «Cache steps as in the shared go.yml» потребует новых шагов; шаг «Release files and samples» — в той же стадии, что `release.yml`, раз он собирает как релиз (R41). Что T66 даёт вызывающему и какие у неё входы — RUN.md, `docs/plan/` и журнал решений `.github` на её теге.
  4. Решение журнала о переходе, поправки к записям, которые он меняет; строки таблиц 3 и 4, которые тег закрывает.
- **Мой шаг:** после мержа — «Rebuild Container»: `.devcontainer/` изменился; оркестратор ждёт этого перед следующей задачей.
- **Как проверить:** проверки частей go и ci; `just -f spec/justfile docs-check-repo` на новом теге; CI pull request зелёный, «Devcontainer» на amd64 и arm64 — тоже; время изменённых job до и после — в отчёте; `/code-review high`, один круг.
- **Разделы:** строки 7.13, 7.14 и 7.18 COORDINATION.md spec; T166 spec; R76 spec, R156 spec; R57, R59, R61.
