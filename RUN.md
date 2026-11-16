# RUN.md — work plan of module-kit

What module-kit does, does not do and has still to decide — [docs/architecture.md](docs/architecture.md); why — [docs/decisions.md](docs/decisions.md). The platform is in `spec/`, a submodule at tag v0.18.6: the component's card — architecture 3.7 spec; the order of repositories — repositories 6 spec. Small tasks, one at a time, gathered in epics; only the current epic is in detail (R19 spec). Other tables, each a file opening with its rules (R153 spec): 2, epics for later — [run/EPICS.md](run/EPICS.md); 3, questions to the author — [run/QUESTIONS.md](run/QUESTIONS.md); 4, debt — [run/DEBT.md](run/DEBT.md). No `run/TASKS.md`: a task outside an epic stands in table 1 (R56, R57).

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

## Table 1. Current epic — E5. Переход на `spec` v0.18.6

| Статус | ID | Задача | Кто | План | Зависит от |
|---|---|---|---|---|---|
| [ ] | T51 | Переход на тег `spec` v0.18.6 | Claude | — | — |

## Epic E5. Переход на `spec` v0.18.6

**Почему эпик такой.** Волна 0 прогона всего «Порядка» COORDINATION.md spec оркестратором (R156 spec; слово автора 2026-11-16) переводит каждый репозиторий на последний тег `spec`, v0.18.6: он несёт итог E14 spec — хук `tools/session` с `NOVATRAIL_RUN_TASKS`, `shutdownAction: "none"` в `devcontainer.json` и правки комментариев Dockerfile заготовки. Текущего эпика не было, а задача не из эпика — первая задача следующего (R56): T51 начала эпик E5, как T45 начала E3 (R57). Номер T51 был свободен. Следующие задачи этого репозитория из «Порядка» идут после перехода.

**Не входит:** код библиотеки и CLI — формат тот же, 0.4; общие проверки `.github` — тот же v0.1.31.

#### T51. Переход на тег `spec` v0.18.6

- **Цель:** `module-kit` стоит на последнем теге `spec`, v0.18.6, — его правила, план и документы по этому тегу, — и остальные задачи эпика прогона COORDINATION.md (волна 0; слово автора 2026-11-16) `spec` идут здесь на нём.
- **Почему:** R156 spec — волна 0 прогона «Сделай эпик E…»: оркестратор первым переводит каждый репозиторий эпика на последний тег `spec`; переход — задача и решение журнала (RULES.md spec, Across repositories). Задача вне эпика или в текущем эпике — по правилам этого RUN.md.
- **Сделать:**
  1. `gh release verify v0.18.6 -R altnovatrail/spec`; подмодуль `spec/` — на v0.18.6; где репозиторий берёт модуль Go или пакет npm `spec` — та же версия, пакет npm — через `release-verify-npm` (R77 spec); тег — везде, где этот репозиторий его называет: CLAUDE.md, README, начало RUN.md, статус `docs/architecture.md`.
  2. Что теги после v0.18.4 до v0.18.6 несут этому репозиторию: `git -C spec log --oneline v0.18.4..v0.18.6`; строки таблицы 4 `spec` с именем репозитория в обратных кавычках — `grep -n '`module-kit`' spec/run/DEBT.md`; карточка architecture 3.7 spec; RULES.md spec. Что из них берёт эта задача, а что — задача своя, решает сессия; отложенное — строкой таблицы 4.
  3. Заготовка `spec/repo-template/` (repositories 2.2 spec): `just new-repo-draft module-kit go 3.7 v0.18.6` на клоне `spec` v0.18.6 во временной папке, как T56 `sim`. `.devcontainer/`, `.claude/`, CLAUDE.md, правила RUN.md и раскладка таблиц плана — по ней, где она изменилась, с отличиями, которые держат решения этого журнала.
  4. Решение журнала о переходе — что тег принёс и что взято —, поправки к записям, которые он меняет; строки таблиц 3 и 4, которые тег закрывает.
  5. Переход больше одного ревью — разбить по правилу 11: эта задача — первая часть, остальные — следующими задачами, и назвать их в отчёте.
- **Как проверить:** `just -f spec/justfile docs-check-repo` на новом теге — зелёный; проверки частей, которые задача меняет, — правило 6 этого RUN.md; `/code-review high`, один круг.
- **Мой шаг:** если задача меняет `.devcontainer/` — назвать это в отчёте: после мержа автор пересобирает контейнер, и оркестратор ждёт этого перед следующей задачей.
- **Разделы:** R156 spec, R76 spec; RULES.md spec; repositories 2.2 spec; карточка architecture 3.7 spec.
