# module-kit — источники

> Внешние факты, на которых стоят документы `docs/` этого репозитория: условия поставщика, цена, закон, стандарт, документация библиотеки. Здесь — их записи: адрес, версия или дата страницы, дата прочтения и цитата на языке источника, — а [architecture.md](architecture.md) и журнал решений держат утверждение и ссылку на запись по её номеру, как на раздел: `sources 1.1` (RULES.md, The bar; R143 spec).
>
> Запись — подраздел «N.M» темы: один источник или одно прочтение нескольких источников ради одного решения. Номер записи не меняется: новая запись встаёт в конец своей темы, новая тема — в конец файла. Источник, прочитанный снова, — пункт со своей датой в той же записи, а прежний текст не правится. Новая версия источника — релиз, тег, коммит, — своя запись: так стоят рядом Go 1.27.3 и 1.27.4 (sources 2.1 и 2.2), общий релиз v0.1.4 и v0.1.10 (sources 3.7 и 3.8), go-yaml v4.0.0-rc.9 и v4.0.0 (sources 4.3 и 4.4). Образец — sources spec. Итоги замеров — не здесь, а в [measurements.md](measurements.md).
>
> До T43 записи стояли пунктами в [architecture.md](architecture.md), раздел 4. Там остался указатель: на этот раздел ссылаются журнал решений и архивы `docs/plan/`, а они не переписываются. Запись начинается с того, для какого решения или раздела её прочитали, так что источники решения находятся по его номеру: `grep -n 'R21' docs/sources.md`.

## 1. MathTrail standalone

### 1.1. Песочница решателей standalone

Для architecture 3.4 — [mathtrail/mathtrail-standalone, `SPEC.md`](https://github.com/mathtrail/mathtrail-standalone/blob/c6cbeb2/SPEC.md), раздел 6 «The Starlark solver», ветка `main`, коммит c6cbeb2, прочитано 2026-10-19:
- 6.1: «the program is written by the same model that wrote the task»;
- 6.4, рекурсия выключена: «A recursive Starlark function recurses on the Go stack, and a stack overflow in Go cannot be recovered: it would take the whole instance down, not the request»;
- 6.5: «Every helper that builds a list is capped at 1,000,000 elements», «a step for every sixteen bytes»; «There is no random number generator, seeded or otherwise. A brute force that samples proves nothing about uniqueness, and a verdict that depends on a seed is a verdict nobody can reproduce from the task alone»;
- 6.6, таблица «The limits»: Steps — 25,000,000; Wall clock — 2 seconds per run; Memory — «What the built-ins keep: 381 MiB at the ceiling above; nine tenths of the instance for the runtime»; Source size — 8 KB; Result — 5 letters; Concurrent runs — 1 per vCPU; Wait for a slot — «3 seconds per run», и запуск, который не дождался места, «does not start, and the task is told it was not checked and cost no attempt». О литерале: «`[{} for i in range(n)]` keeps 512 bytes for every nine steps — 2.6 million dictionaries at the ceiling, and an instance of 1 GiB was killed for its memory by them». О калибровке: «The bench measured the worst of the 450 at 2,081,362 steps, which is what moved the ceiling from 10,000,000 to 25,000,000» (R53 standalone) и «the worst of all 603 is 2,413,617 steps». О дыре памяти: «a solver written that way takes the instance down rather than the request» и «The fix, if the gap ever bites, is a solver in a process of its own with a memory limit on it — a change of shape rather than a setting, and not one v1 pays for».

### 1.2. Дыра памяти шире, чем в SPEC

Для architecture 3.4 — [`docs/decisions.md`](https://github.com/mathtrail/mathtrail-standalone/blob/c6cbeb2/docs/decisions.md) standalone, R53 standalone, коммит c6cbeb2, прочитано 2026-10-19: «the methods of the values themselves take a lazy range of any length in one step — `x.extend(range(3000000))` is nineteen steps», «so no cap on the helpers can close the gap, and only a limit on the memory of the process can»; о времени: «the clock is seen only between instructions» и «a process of its own would bound its time as well as its memory».

### 1.3. Цена в байтах, места и память экземпляра

Для architecture 3.4 — R125 standalone, тот же `docs/decisions.md`, коммит c6cbeb2, прочитано 2026-10-19: «an instance runs one solver for each vCPU, and it has 1 GiB, nine tenths of it told to the runtime»; отвергнуто: «Four slots with the clock as insurance. Verdicts depend on the load» и «The process of its own now. A change of shape, and the author's decision»; открыто: «One account handing in solvers that spend their whole clock, for requests nobody opened, keeps an instance's single slot busy» и «Such a hand-in spends no attempt, so the account pays nothing for it. That stays open».

### 1.4. Интерпретатор Starlark standalone

Для architecture 3.4 — [`go.mod`](https://github.com/mathtrail/mathtrail-standalone/blob/c6cbeb2/go.mod), коммит c6cbeb2, прочитано 2026-10-19: `go.starlark.net v0.0.0-20260908191801-89a6a09411d5`.

### 1.5. Случайность в коде standalone

Для architecture 3.3 — поиск по `internal/` и `cmd/` того же коммита, 2026-10-19:
- вне тестов `math/rand/v2` импортируют только `internal/store/drive/calls.go` и `internal/transport/mcp/afresh.go`, и оба вызова дают случайную длительность паузы (`time.Duration`); генератора задач от зерна там нет;
- в тестах генераторы PCG получают зерно числами, без хеша: `rand.NewPCG(uint64(grade), uint64(child))` в `internal/domain/tutor/simulation_test.go`, `rand.NewPCG(7, 7)` в `pause_test.go`, `rand.NewPCG(d.Dice, 0)` в `property_test.go` и другие.

### 1.6. Вывод зёрен в стенде standalone

Для architecture 3.3 — [`tools/learners/child.go`](https://github.com/mathtrail/mathtrail-standalone/blob/c6cbeb2/tools/learners/child.go), коммит c6cbeb2, прочитано 2026-10-19: функция `seeded(unit, purpose string) *rand.Rand` с комментарием «seeded is the random stream of one unit of the experiment for one purpose»; она пишет в `fnv.New64a()` строку `"%d|%s|%s|%s"` из `masterSeed`, `experiment`, `unit` и `purpose` и возвращает `rand.New(rand.NewPCG(hash.Sum64(), masterSeed))` из `math/rand/v2`. Та же функция есть в `research/experiments/` того же коммита.

## 2. Go и его тулчейн

### 2.1. Go 1.27.3

Для R04 — список релизов go.dev, `https://go.dev/dl/?mode=json`, прочитано 2026-10-21: новейший стабильный релиз — `go1.27.3`; sha256 архива `go1.27.3.linux-amd64.tar.gz` — `65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40`, `go1.27.3.linux-arm64.tar.gz` — `a06e99309dbcbdca1a8ccd2282c3ee146e55930886e26b4f1193269587165f49`. Оба архива скачаны и сверены `sha256sum` 2026-10-21.

### 2.2. Go 1.27.4

Для R54 — список релизов go.dev, `https://go.dev/dl/?mode=json&include=all`, прочитано 2026-11-12: sha256 архива `go1.27.4.linux-amd64.tar.gz` — `7226736c2339719cda28fbf3a54d71aa63bcefe376e98cef00034268dab24840`, `go1.27.4.linux-arm64.tar.gz` — `3392582988a273a87ab5043ee1b2c9c17a2c3c00580bf9326c47a2e0aaaeac92`, те же, что в `.devcontainer/Dockerfile` `spec` v0.16.3. Архив amd64 скачан и сверен `sha256sum` 2026-11-12.

### 2.3. `GOTOOLCHAIN=local`

Для R04 — [Go Toolchains](https://go.dev/doc/toolchain), прочитано 2026-10-21: «When GOTOOLCHAIN is set to local, the go command always runs the bundled Go toolchain».

### 2.4. JSON в Go 1.27.3

Для R09 — исходники в `$(go env GOROOT)` стадии `toolchain`, прочитано 2026-11-02:
- `src/internal/buildcfg/exp.go`, строка 87: `JSONv2: true` в наборе `baseline`, то есть эксперимент `jsonv2` включён по умолчанию. Файлы `encoding/json/jsontext` начинаются с `//go:build goexperiment.jsonv2`. `go list encoding/json/jsontext` без флагов находит пакет, а с `GOEXPERIMENT=nojsonv2` отвечает «build constraints exclude all Go files»;
- `go doc encoding/json/jsontext.Value.Canonicalize`: «Canonicalize canonicalizes the raw JSON value according to the JSON Canonicalization Scheme (JCS) as defined by RFC 8785» и «JSON numbers are formatted as double precision numbers»;
- `encoding/json` в go1.27 собран из файлов `v2_*.go` с `//go:build goexperiment.jsonv2`, а `encode.go` — с `//go:build !goexperiment.jsonv2`. В `internal/jsonflags/flags.go` `DefaultV1Flags` включает `EscapeForHTML` и `EscapeForJS`, а в `internal/jsonwire/encode.go`, строка 97, — `case (r == '\u2028' || r == '\u2029') && flags.Get(jsonflags.EscapeForJS):`. `SetEscapeHTML(false)` снимает только `EscapeForHTML`.

### 2.5. `math/rand/v2` и `hash/fnv` в Go 1.27.3

Для R10 — исходники в `$(go env GOROOT)` стадии `toolchain`, прочитано 2026-11-02:
- `src/math/rand/v2/pcg.go`: `mulHi = 2549297995355413924`, `mulLo = 4865540595714422341`, `incHi = 6364136223846793005`, `incLo = 1442695040888963407` и выход DXSM с `cheapMul = 0xda942042e4dd58b5` — те же числа, что в [тексте формата](../spec/format/README.md), раздел 22;
- `src/math/rand/v2/rand.go`: у `Uint64N` — «It panics if n == 0», у `Int64N` — «It panics if n <= 0», у `Shuffle` — «Shuffle panics if n < 0». `Perm` о панике не говорит, но `Perm(-1)` падает раньше `Shuffle`, в `make([]int, n)`: «panic: runtime error: makeslice: len out of range», строка 235 (проба, 2026-11-02). На 32 битах `uint64n` идёт через `uint32n`, и комментарий к нему говорит: «On 32-bit systems, here we need to implement that same logic in 32-bit math, both to preserve the exact output sequence observed on 64-bit machines and to preserve the optimization that the unbiasing loop almost never runs». О неизменности потоков документация пакета не говорит: `go doc -all math/rand/v2` такой фразы не находит;
- `src/math/rand/v2/regress_test.go`: «Test that random number sequences generated by a specific seed do not change from version to version» и «Do NOT make changes to the golden outputs. If bugs need to be fixed in the underlying code, find ways to fix them that do not affect the outputs»;
- `src/hash/fnv/fnv.go`: «Package fnv implements FNV-1 and FNV-1a, non-cryptographic hash functions», `offset64 = 14695981039346656037` и `prime64 = 1099511628211`.

### 2.6. Обещание Go о потоках

Для R10 — блог Go, [«Evolving the Go Standard Library with math/rand/v2»](https://go.dev/blog/randv2), 1 мая 2024, прочитано 2026-11-02. О `math/rand`: «At that point, we decided Go 1 compatibility included the specific random outputs for a given seed and added a test». О `math/rand/v2` — после того, как `IntN` и родственные функции получили метод Лемира: «The concrete rand.Rand API is now locked in to that value stream, so we will not be able to take advantage of any optimizations yet to be discovered».

### 2.7. Тулчейн главного модуля

Для R10 — [Go Toolchains](https://go.dev/doc/toolchain), прочитано 2026-11-02: «The choice of Go toolchain being used depends on the GOTOOLCHAIN environment setting and the go and toolchain lines in the main module's go.mod file or the current workspace's go.work file» и «The go line for each module sets the language version the compiler enforces when compiling packages in that module».

### 2.8. Модули `spec` у прокси Go

Для R20 — `https://proxy.golang.org/github.com/altnovatrail/spec/@v/v0.9.2.info` и `…/spec/contracts/@v/v0.9.2.info`, прочитано 2026-11-04: у обоих `"Hash":"6f86ff50c76dacb79b7faf673713ddf02e42f531"`, а `Ref` — `refs/tags/v0.9.2` и `refs/tags/contracts/v0.9.2`. `go help mod download`, Go 1.27.3: у `-json` есть поле `Origin any // provenance of module`; `go mod download -json <модуль>@<версия>` печатает его, а без версии — нет (проба, 2026-11-04).

### 2.9. Покрытие из чужих тестов

Для R20 — `go help testflag`, Go 1.27.3, прочитано 2026-11-04: «-coverpkg pattern1,pattern2,pattern3 Apply coverage analysis in each test to packages whose import paths match the patterns. The default is for each test to analyze only the package being tested».

## 3. Проверки, CI и релиз

### 3.1. Инструменты

Для R04 — прокси модулей Go, `https://proxy.golang.org/<модуль>/@latest`, прочитано 2026-10-21: `golang.org/x/vuln` — v1.9.0 (2026-10-11), `github.com/google/go-licenses/v2` — v2.0.1 (2025-09-08). Лицензии — API GitHub, `repos/<владелец>/<репозиторий>/license`, 2026-10-21: golang/vuln — BSD-3-Clause, google/go-licenses — Apache-2.0.

### 3.2. `go-licenses check`

Для R04 — [README v2.0.1](https://github.com/google/go-licenses/blob/v2.0.1/README.md), раздел Usages, Check, прочитано 2026-10-21: «Allow only specific license names» — `go-licenses check <package> [package...] --allowed_licenses=<comma separated license names>`. Проба 2026-10-21: со списком `spec` рецепт `ci-licenses` проходит, а со списком из одной MIT отвечает «Not allowed license 'Apache-2.0' found for library 'github.com/altnovatrail/module-kit/cmd/module-kit'».

### 3.3. Образец рецептов

Для R04 — [`justfile` canvas-api](https://github.com/MathTrail/canvas-api/blob/0901af0439280334b3da33ffc4cc5c853e4403ff/justfile), коммит 0901af0, прочитано 2026-10-21: `ci-lint`, `ci-test`, `ci-vuln` и `ci-licenses`, а об инструментах — «Each is a Go program run straight from its module, so none of them is installed anywhere».

### 3.4. SonarCloud, docker:S8545

Для R04 — правило «Go dependencies should be locked to verified versions» (API SonarCloud, `api/rules/show?key=docker:S8545`, 2026-10-21). На pull request #8 оно отметило строку `go install` gopls в Dockerfile: «Dependency versions are not predictable. Use a lock-file enforcing command instead». У canvas-api такие находки открыты у двух строк его Dockerfile, 198 и 199, хотя там стоит `-mod=readonly` (API SonarCloud, проект MathTrail_canvas-api, 2026-10-21).

### 3.5. Autobuild Go в CodeQL

Для R06 — [GitHub Docs, «CodeQL build options and steps for compiled languages»](https://docs.github.com/en/code-security/reference/code-scanning/codeql/build-options-for-compiled-languages), раздел «Autodetection for Go», исходник в `github/docs` на коммите c008b66, прочитано 2026-10-21: «Invoke `make`, `ninja`, `./build` or `./build.sh` (in that order) until one of these commands succeeds and a subsequent `go list ./...` also succeeds»; «By default, test code (code in files ending in `_test.go`) is not analyzed».

### 3.6. Лицензии в релизе

Для R07:
- Apache-2.0 — [LICENSE](../LICENSE), раздел 4 (a): «You must give any other recipients of the Work or Derivative Works a copy of this License»;
- лицензия Go — `LICENSE` в корне Go 1.27.3 стадии `toolchain`, `$(go env GOROOT)/LICENSE`, прочитано 2026-10-21: «Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution». Бинарник Go несёт стандартную библиотеку, а сторонних модулей у `module-kit` нет (`go.mod`).

### 3.7. Общий релиз

Для R07 — [`release.yml`](https://github.com/altnovatrail/.github/blob/11435dbe61d3d4885c9a8dd30686e8111e26cabb/.github/workflows/release.yml) и R18 .github, релиз `.github` v0.1.4, коммит 11435db, прочитано 2026-10-21: условия рецепту `release-files` — стадия `toolchain` выпускаемого коммита, его дерево из `git archive`, запуск под uid раннера с `HOME=/tmp` и без токенов, в `dist/` — только простые файлы с именами по `^[A-Za-z0-9_+][A-Za-z0-9._+-]*$` и без `.sigstore.json`. Релиз неизменяемый: `gh release verify v0.1.4 -R altnovatrail/.github` — «Release v0.1.4 verified!».

### 3.8. Общий релиз v0.1.10

Для R07 и T08 — сравнение `release.yml` `.github` v0.1.4 (11435db) и v0.1.10 (f3addb7), `gh api repos/altnovatrail/.github/compare/…`, прочитано 2026-11-02. Рецепты `spec` релиз запускает скриптом `spec/tools/toolchain/toolchain.sh` из подмодуля вызывающего, а до сборки проверяет: «that spec/ be a submodule of altnovatrail/spec, checked out at the commit the caller pins and at a tag of spec» и «that the tag hold the script, which spec ships from v0.2.17; anything else stops it before anything is built, with the reason». Сумму скрипта он сверяет и после сборки: «spec/tools/toolchain/toolchain.sh changed in the copy of the checkout while the release ran». Условия рецепту `release-files` выше прочитаны на v0.1.4: на v0.1.10 их проверит первый релиз.

### 3.9. Покрытие и перечень `paths`

Для R18 — [docs/architecture.md `.github`](https://github.com/altnovatrail/.github/blob/89d3ceee0b9977b3334488a4133d084d76474e0d/docs/architecture.md), раздел 3.3 — «Покрытие», job «Go» и перечень вида `go`, — и [`go.yml`](https://github.com/altnovatrail/.github/blob/89d3ceee0b9977b3334488a4133d084d76474e0d/.github/workflows/go.yml), релиз `.github` v0.1.15, коммит 89d3cee, прочитано 2026-11-04. Из общих workflow между v0.1.10 и v0.1.15 менялись только `go.yml`, `release.yml` и `typescript.yml` (`gh api repos/altnovatrail/.github/compare/v0.1.10...v0.1.15`, 2026-11-04).

### 3.10. gopls из lock-файла

Для R19 — R12 tutor и его [`.devcontainer/Dockerfile`](https://github.com/altnovatrail/tutor/blob/be607e56f87bf94a93b1d318f9cd578a3e3c221e/.devcontainer/Dockerfile) и `.devcontainer/gopls/`, `main` be607e5, прочитано 2026-11-04. Находок docker:S8545 у `tutor` нет (API SonarCloud, `api/issues/search?componentKeys=altnovatrail_tutor&rules=docker:S8545`, 2026-11-04). Его Scorecard Pinned-Dependencies отмечает в Dockerfile только `npmCommand not pinned by hash` (API GitHub, `code-scanning/alerts`, 2026-11-04). gopls v0.24.0 от 2026-10-14 — последний релиз (`https://proxy.golang.org/golang.org/x/tools/gopls/@latest`, прочитано 2026-11-04); лицензия golang/tools — BSD-3-Clause (API GitHub, `repos/golang/tools/license`, 2026-11-04).

### 3.11. Когда Dependabot предлагает версию

Для поправки 2026-11-04 к R18 — GitHub Docs, прочитано 2026-11-04: [«Configuring Dependabot version updates»](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuring-dependabot-version-updates) — «When this file is checked in, Dependabot checks the manifest files on the default branch for outdated dependencies»; [«Dependabot options reference»](https://docs.github.com/en/code-security/dependabot/working-with-dependabot/dependabot-options-reference), `cooldown` — «A new version is not considered for a version update until 3 days after its release», а с какого мгновения и как считаются дни, там не сказано. Релиз `.github` v0.1.15 опубликован 2026-11-02 в 03:40:33 UTC (`gh release view v0.1.15 -R altnovatrail/.github`, 2026-11-04).

## 4. Библиотеки Go

### 4.1. Библиотеки JCS для Go

Для R09 — прокси модулей Go и API GitHub, прочитано 2026-11-02:
- `github.com/gowebpki/jcs` — `@latest` v1.0.2 от 2026-09-21, лицензия Apache-2.0. В `go.mod` — `require github.com/stretchr/testify v1.12.1`. Issue #5 «RFC 8785: reversed surrogate pair becomes U+FFFD and invalid UTF-8 passes through» закрыт 2026-09-20. Число он читает как double: `jcs.go` на теге v1.0.2, строки 530 и 539, — `strconv.ParseFloat(value, 64)`, затем `NumberToJSON(ieeeF64)`;
- `github.com/cyberphone/json-canonicalization` — тегов версий нет, `@latest` — псевдоверсия `v0.0.0-20261011152517-32a48195b9b2`. Лицензия по API GitHub — `NOASSERTION`, `go.mod` в репозитории нет. Issue #33 «Go import (as documented) broken» открыт с 2025-03-11.

### 4.2. go.starlark.net

Для R11 — [google/starlark-go](https://github.com/google/starlark-go), лицензия BSD-3-Clause (API GitHub), прочитано 2026-11-02:
- тегов версий нет. `https://proxy.golang.org/go.starlark.net/@latest` — `v0.0.0-20261021141710-4f4267fb140d`, коммит 4f4267fb140d от 2026-10-21; standalone закрепил `v0.0.0-20260908191801-89a6a09411d5` (`go.mod` standalone, коммит 89986bc);
- pull request #663 «starlark: check steps/cancellation only on back-edges», слит 2026-10-05 коммитом de228787c8fb: «We only need to increment the step counter and check for cancellation in loops (back edges) and calls. We increment the counter by a constant scaled by the number of instructions between such checks» и «Clients may need to adjust their SetMaxSteps parameter to taste». В `starlark/interp.go` на 4f4267fb140d — `const stepCost = 10 // empirical average number of instructions per back-edge or call`, а на 89a6a09411d5, строка 109, — `thread.Steps++` на каждой инструкции;
- `starlark/eval.go` на 4f4267fb140d. О `Cancel`: «There may be a delay before the interpreter observes the cancellation if the thread is currently in a call to a built-in function». О повторе списка и строки: «It's always possible to overeat in small bites but we'll try to stop someone swallowing the world in one gulp» и `const maxAlloc = 1 << 30`;
- `starlark/interp.go` на 4f4267fb140d о рекурсии. При выключенной: «We look for the same function code, not function value, otherwise the user could defeat the check by writing the Y combinator». При включённой — `if len(thread.stack) > 100_000`, с pull request #619 от 2025-10-27: «Exceeding this limit should not crash the Go program»;
- pull request #672, слит 2026-10-21: «calling one of them without an argument dereferenced a nil `Iterable` and panicked, e.g. `set([1]).difference()`».

### 4.3. go-yaml v4

Для R21 — прокси модулей Go, `https://proxy.golang.org/go.yaml.in/yaml/v4/@latest`, прочитано 2026-11-04: `v4.0.0-rc.9` от 2026-11-02, коммит 2141fcc, тег `refs/tags/v4.0.0-rc.9`. Лицензия по API GitHub (`repos/yaml/go-yaml/license`) — Apache-2.0, а `NOTICE` модуля: «The following files were ported to Go from C files of libyaml, and thus are still covered by their original MIT license». `go.mod` модуля — `go 1.18`, без `require`. README на том же теге: «v4.0.0 *is not released yet.* It is intended to be released in October 2026», «v2 and v3 remain stable» и «The library supports most of YAML 1.2 and retains selected YAML 1.1 behavior».

### 4.4. go-yaml v4.0.0

Для R45 — `https://proxy.golang.org/go.yaml.in/yaml/v4/@latest`, прочитано 2026-11-06: `v4.0.0` от 2026-11-04T18:36:35Z, коммит f2f9563, тег `refs/tags/v4.0.0`. От rc.9 — пять коммитов, `gh api repos/yaml/go-yaml/compare/v4.0.0-rc.9...v4.0.0`; `yts/known-failing-tests` v4.0.0 по-прежнему называет 2SXE и W5VH.

### 4.5. Валидатор JSON Schema

Для R23 — прочитано 2026-11-04:
- `github.com/santhosh-tekuri/jsonschema/v6` — `https://proxy.golang.org/github.com/santhosh-tekuri/jsonschema/v6/@latest`: `v6.0.3` от 2026-06-28, коммит b0fc661, `refs/tags/v6.0.3`. Лицензия по API GitHub (`repos/santhosh-tekuri/jsonschema/license`) — Apache-2.0. Его `go.mod` — `go 1.21`, `require github.com/dlclark/regexp2 v1.11.0 // used for testing` и `golang.org/x/text v0.14.0`. Bowtie, [bowtie.report](https://bowtie.report), прогон 2026-11-04 06:21 UTC: у draft 2020-12 go-jsonschema (v6.0.2) проходит 1300 тестов из 1300;
- его исходник на v6.0.3, `validator.go`: у `propertyNames` — `verr.InstanceLocation = vd.vloc`, без копии, а `instanceLocation()` других ошибок — `slices.Clone(vd.vloc)`. Issue [#276](https://github.com/santhosh-tekuri/jsonschema/issues/276) «propertyNames and contentSchema errors: InstanceLocation is overwritten by later siblings» открыт с 2026-09-24; pull request #278 закрыт без мержа 2026-10-02, «Closing in favor of earlier community PR #269», а #269 открыт. Проба 2026-11-04 на `spec.Format` v0.9.2: `examples/task/invalid/parameter-named-if.yaml` получил `propertyNames at #/items/0` вместо `#/template/parameters`. Коммит 9ae86900d от 2026-10-31, после v6.0.3, — «validator/fix: big.Rat fails to parse large exponents causing panic»: до него `numValidate` читает число `new(big.Rat).SetString(fmt.Sprintf("%v", v))` и результат не проверяет. Как словарь включается — `Compiler.RegisterVocabulary` и `AssertVocabs`, а ошибку расширения даёт `ValidatorContext.AddError` (`vocab.go`);
- `golang.org/x/text` — `@latest` `v0.43.0` от 2026-10-09; govulncheck v1.9.0 на v0.14.0: GO-2026-6629 «Panic parsing crafted input in x/text/secure/precis», исправлена в v0.41.0, и GO-2026-5970 «Infinite loop on invalid input», исправлена в v0.39.0, — «your code doesn't appear to call these vulnerabilities»; на v0.43.0 — «No vulnerabilities found». go-licenses v2.0.1: santhosh — Apache-2.0, x/text — BSD-3-Clause;
- другие библиотеки: `github.com/kaptinlin/jsonschema` — `v0.9.10` от 2026-09-13, MIT, Bowtie его не проверяет; `github.com/google/jsonschema-go` — `v0.4.3` от 2026-04-17, MIT, валидация останавливается на первой ошибке; Corvus Go (`corvus-dotnet/Corvus.JsonSchema`, `src-go/`) — v0.1.4 от 2026-11-03, первый релиз v0.1.0 от 2026-10-08, Apache-2.0, шаблоны — по ECMA-262; `xeipuuv/gojsonschema` v1.2.0 — draft-04, -06 и -07; `qri-io/jsonschema` v0.2.1 — draft 2019-09.

### 4.6. go.starlark.net в песочнице

Для R28 — исходники `go.starlark.net` `v0.0.0-20261021141710-4f4267fb140d` из кэша модулей, прочитано 2026-11-04:
- `starlark/library.go`, у `Universe`: «The Go application may add or remove items from the universe dictionary before Starlark evaluation begins»; в нём 33 имени — те, что перечисляет [текст формата](../spec/format/README.md), раздел 24;
- `starlark/eval.go`: у `Thread.Print` — «If nil, fmt.Fprintln(os.Stderr, msg) is used instead»; у `OnMaxSteps` — «The default behavior is to call thread.Cancel("too many steps")»; `Cancel` — «Atomically set cancelReason, preserving earlier reason if any»; у `EvalExprOptions` — «Resolving an expression mutates it. Do not call EvalExprOptions more than once for the same expression», а `makeExprFunc` резолвит имена `resolve.ExprOptions(opts, expr, env.Has, Universe.Has)`;
- `starlark/interp.go`: `checkSteps` — `thread.Steps += stepCost` и `thread.Steps >= thread.maxSteps`, а `stepCost = 10`;
- `starlark/int_posix64.go`, сборка для linux и darwin на 64 битах: «It reserves a 4GB portion of the address space using mmap», `unix.Mmap(-1, 0, len, unix.PROT_READ, unix.MAP_PRIVATE|unix.MAP_ANON)`, а при ошибке — `log.Printf("Starlark failed to allocate 4GB address space: %v. Integer performance may suffer.", err)`;
- `go.mod` модуля — `go 1.25.0` и `golang.org/x/sys v0.42.0` среди `require`; лицензия google/starlark-go по API GitHub — BSD-3-Clause.

### 4.7. `golang.org/x/sys`

Для R28 — прочитано 2026-11-04: `https://proxy.golang.org/golang.org/x/sys/@latest` — `v0.49.0` от 2026-10-06; лицензия golang/sys по API GitHub — BSD-3-Clause. govulncheck v1.9.0 на v0.42.0: GO-2026-5024 «Invoking integer overflow in NewNTUnicodeString in golang.org/x/sys/windows», исправлена в v0.44.0, платформа windows, — «your code doesn't appear to call these vulnerabilities»; на v0.49.0 — «No vulnerabilities found». go-licenses v2.0.1: go.starlark.net и `golang.org/x/sys/unix` — BSD-3-Clause.

### 4.8. protobuf-go

Для R29 — исходники `google.golang.org/protobuf` v1.36.12 из кэша модулей, прочитано 2026-11-05:
- `internal/errors/errors.go`, `InvalidUTF8`: «field %v contains invalid UTF-8» — так `proto.Marshal` отказывает строке не в UTF-8, а `TestNotUTF8IsNoMessage` держит это на наблюдении;
- `internal/filedesc/desc.go`, `EnforceUTF8`: «If this method does not exist, the default is to enforce valid UTF-8»;
- `encoding/protojson/decode.go`, `unmarshalEnum`: число JSON — `protoreflect.EnumNumber(n)`, если оно влезает в 32 бита, без поиска в перечислении. Так `protojson.Unmarshal` читает причину вектора `1000`;
- лицензия — `LICENSE` модуля, BSD-3-Clause, как его называет go-licenses v2.0.1 (`ci-licenses`).

## 5. Стандарты и спецификации

### 5.1. Semantic Versioning 2.0.0

Для R13 и architecture 3.5 — [semver.org/spec/v2.0.0.html](https://semver.org/spec/v2.0.0.html), прочитано 2026-11-02:
- пункт 3: «Once a versioned package has been released, the contents of that version MUST NOT be modified»;
- пункт 4: «Major version zero (0.y.z) is for initial development. Anything MAY change at any time»;
- пункт 8: «Major version X (X.y.z | X > 0) MUST be incremented if any backward incompatible changes are introduced to the public API. It MAY also include minor and patch level changes» — часть, найденную сравнением, автор может поднять выше.

### 5.2. Cargo о 0.y.z

Для R13 и architecture 3.5 — [The Cargo Book, «SemVer Compatibility»](https://doc.rust-lang.org/cargo/reference/semver.html), прочитано 2026-11-02: «Initial development releases starting with "0.y.z" can treat changes in "y" as a major release, and "z" as a minor release. "0.0.z" releases are always major changes. This is because Cargo uses the convention that only changes in the left-most non-zero component are considered incompatible». Так же — oasdiff: «Below `1.0.0`, semver gives the minor version the major's role» ([`docs/VERSIONING.md`](https://github.com/oasdiff/oasdiff/blob/2daca9b/docs/VERSIONING.md), коммит 2daca9b, прочитано 2026-11-02).

### 5.3. Инструменты, которые проверяют semver по содержимому

Для R13, прочитано 2026-11-02:
- gorelease — [pkg.go.dev/golang.org/x/exp/cmd/gorelease](https://pkg.go.dev/golang.org/x/exp/cmd/gorelease): «If no version is proposed with -version, gorelease suggests the lowest version consistent with semantic versioning». Находит наименьшую часть и предлагает версию;
- apidiff — [pkg.go.dev/golang.org/x/exp/apidiff](https://pkg.go.dev/golang.org/x/exp/apidiff): «No tool can detect behavioral changes; and even if it could, whether a behavioral change is a breaking change or not depends on many factors, such as whether it closes a security hole or fixes a bug». У модуля поведение — проверка задачи, и её сравнение видит;
- cargo-semver-checks, лицензия Apache-2.0 — [README](https://github.com/obi1kenobi/cargo-semver-checks/blob/161450e/README.md), коммит 161450e: «By default, `cargo-semver-checks` uses crates.io to look up the previous version of the crate, which is used as the baseline for semver-checking the current version of the crate», а прежнюю версию можно дать и папкой, `--baseline-root`. И: «A design goal of `cargo-semver-checks` is to not have false positives». Требует не меньше найденного, а автор задаёт версию сам;
- oasdiff — тот же `docs/VERSIONING.md`: «a breaking change requires a major version increase», а под «What is not checked» — «Which version bump a non-breaking change deserves. Only breaking changes are checked»;
- `elm bump` — [`terminal/src/Main.hs`](https://github.com/elm/compiler/blob/1c278bf/terminal/src/Main.hs) elm/compiler, коммит 1c278bf: «so there cannot be MAJOR changes hiding in PATCH releases in Elm!»; правило — [`builder/src/Deps/Diff.hs`](https://github.com/elm/compiler/blob/1c278bf/builder/src/Deps/Diff.hs), строки 308–312: `addMag = if null added then M.PATCH else M.MINOR`, `removeMag = if null removed then M.PATCH else M.MAJOR` и наибольшее из частей. Удаление — мажорная, добавление — минорная, и Elm требует ровно найденную версию;
- версия данных — рецепт Frictionless Data Package Version, [specs.frictionlessdata.io/patterns](https://specs.frictionlessdata.io/patterns/): мажорная — «Combine, split, delete or change the meaning of data that is referenced by another data resource», патч — «Correct errors in existing data». Часть здесь объявляет автор, а инструмента нет;
- [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/): «fix: a commit of the type fix patches a bug in your codebase (this correlates with PATCH in Semantic Versioning)». Часть объявляет автор — типом коммита, — а содержимого инструмент не смотрит.

### 5.4. YAML 1.2.2

Для R21 — [yaml.org/spec/1.2.2](https://yaml.org/spec/1.2.2/), прочитано 2026-11-04:
- 5.1: «To ensure JSON compatibility, YAML processors must allow all non-C0 characters inside quoted scalars»;
- 5.2: «Byte order marks may appear at the start of any document» и «To allow for JSON compatibility, byte order marks are also allowed inside quoted scalars»; `nb-char ::= c-printable - b-char - c-byte-order-mark`;
- 6.1: «To maintain portability, tab characters must not be used in indentation»;
- 6.3: `s-flow-line-prefix(n) ::= s-indent(n) s-separate-in-line?`;
- 7.4.2: «the ":" indicator must appear at most 1024 Unicode characters beyond the start of the key»;
- 8.1.1.1: «It is an error for any of the leading empty lines to contain more spaces than the first non-empty line»;
- 8.2.2: неявные ключи «are limited to a single line and must not span more than 1024 Unicode characters».

### 5.5. Набор тестов YAML

Для R21 — [yaml/yaml-test-suite](https://github.com/yaml/yaml-test-suite), релиз данных `data-2022-01-17`, коммит 5f49729, его же берёт v4 (`YTS-TAG ?= data-2022-01-17` в `GNUmakefile` v4.0.0-rc.9); лицензия — MIT, файл `License` на теге `v2022-01-17`, коммит 6d48918 (API GitHub, прочитано 2026-11-04). В релизе 402 случая, из них 94 с файлом `error`. v4 rc.9 на них, без проходов `internal/yaml`: пропускает пять неверных — 9C9N, QB6E, DK95/01, Y79Y/003 и S98Z, — и отвергает 14 верных, из них 12 — директивы и якоря.

### 5.6. Сигнатуры медиафайлов

Для R22 — [WHATWG MIME Sniffing](https://mimesniff.spec.whatwg.org/), прочитано 2026-11-04: 6.2.3 «Signature for MP3 without ID3», «To match an mp3 header», шаг 1 — «If length is less than 4, return false». Так же — `isMp3Header` в `spec/tools/format-check/module.mjs` на v0.9.2: `bytes.length >= 4`.

### 5.7. Unicode нормализации

Для R27 — прочитано 2026-11-04:
- Go 1.27.3, `src/unicode/tables.go`: `const Version = "17.0.0"`; `CaseRanges` — «the table describing case mappings for all letters with non-self mappings», и `unicode.ToLower` по нему — простое отображение: проба даёт для U+0130 знак U+0069, один, а U+0085 — White_Space, U+FEFF — нет;
- `golang.org/x/text` v0.43.0, `unicode/norm`: `tables17.0.0.go` — `//go:build go1.27`, `Version = "17.0.0"`, а `tables15.0.0.go` — `//go:build !go1.27`, `Version = "15.0.0"`; так же `cases` и `rangetable`. `normalize.go`, `decomposeSegment`: «It inserts 0x034f (Grapheme Joiner) when it encounters a sequence of more than 30 non-starters», а `composition.go` — `maxNonStarters = 30`. Проба: `"a"` и 31 знак U+0301 — на выходе `norm.NFKC` есть U+034F;
- [UAX #15](https://www.unicode.org/reports/tr15/), раздел 13 «Stream-Safe Text Format», версия 18.0.0, ревизия 58, от 2026-08-12: «The Stream-Safe Text Process is the process of producing a Unicode string in Stream-Safe Text Format by processing that string from start to finish, inserting U+034F COMBINING GRAPHEME JOINER (CGJ) within long sequences of non-starters» и «The Stream-Safe Text Process preserves all of the four normalization forms defined in this annex (NFC, NFD, NFKC, NFKD). However, normalization and the Stream-Safe Text Process do not commute»;
- вторая реализация `spec` — `spec/tools/format-check/verdicts.mjs` на v0.9.2: NFKC — `text.normalize("NFKC")` Node 24.21, ICU 78.3, Unicode 17.0 (R56 spec), без Stream-Safe Text Process; строчная буква — `toLowerCase` с особым случаем U+0130, «as Go's unicode.ToLower».

### 5.8. Строки proto3 и ProtoJSON

Для R29 — прочитано 2026-11-05:
- [Language Guide (proto 3)](https://protobuf.dev/programming-guides/proto3/), таблица «Scalar Value Types», `string`: «A string must always contain UTF-8 encoded or 7-bit ASCII text, and cannot be longer than 2³²»;
- [ProtoJSON Format](https://protobuf.dev/programming-guides/json/), строка `enum`: «The name of the enum value as specified in proto is used. Parsers accept both enum names and integer values».

## 6. Вторая реализация `spec`

### 6.1. Вторая реализация `spec`

Для R21 — `spec/tools/format-check/yaml.mjs` на v0.9.2, npm `yaml` 2.9.1, проба 2026-11-04 через её `readModuleFile`: `a: x<U+2028>  y` она читает как `"x\u2028  y"`; `%YAML 1.2` — `yaml.no-directives`; BOM после `...` — `yaml.encoding`; `&x.y` — `yaml.no-anchors`; пустой ключ `? ` дважды — `yaml.plain-strings` дважды и `yaml.unique-keys`; «не YAML» — `\U00110000`, управляющий знак в комментарии, неявный ключ длиннее 1024 знаков и недоотступленные строки продолжения. Файл с переводами строки CR она отвергает, `...` в начале потока считает документом, а U+FEFF в простом и блочном скаляре и в комментарии принимает.

### 6.2. Вторая реализация `spec` и схемы

Для R23 — `spec/tools/format-check/` на v0.9.2, прочитано 2026-11-04: `schemas.mjs`, строки 368–375, — `new Ajv2020({allErrors: true, validateFormats: false, …})` и `ajv.addKeyword("x-open")`; `cases.mjs`, строки 144–156, — пример проходит, если `errors.some((error) => error.keyword === header.keyword && error.instancePath === header.pointer)`, а `expectedError`, строка 828, раскодирует место `decodeURIComponent`; `cases.mjs`, строки 219–239, — `module.file-ids` «read only from a file that passed its schema, so that a file the schema refuses breaks that and nothing more», а правила целого модуля — `if (values.partsOk)`. Открытый вид строит `openview.mjs` своим обходом, а ajv отвечает только, проходит ли подсхема (`passes`, строки 116–121).
