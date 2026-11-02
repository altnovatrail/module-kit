# module-kit — архитектура

> Статус: каркас из заготовки `spec`, 2026-10-19; сверен с `spec` v0.2.11 2026-10-20 (R03), с v0.2.13 2026-10-21 (R05) и с v0.6.3 2026-11-02 (R08); первый код — заготовка CLI, 2026-10-21 (R04); релизы — общим `release.yml` `.github`, с 2026-10-21 (R07).
>
> Что делает компонент, чего не делает и что о нём ещё предстоит решить. Это подробности его карточки — architecture 3.7 spec. Карточка остаётся сводкой: если подробности меняют то, что в ней сказано, она правится в `spec` тем же путём, что и вопросы в его реестры (RULES.md, Across repositories).
>
> Почему решено так, а не иначе, — в [decisions.md](decisions.md); план работ — в [RUN.md](../RUN.md). Понятия — concepts spec. Документы и решения `spec` лежат в `spec/docs/` и цитируются с суффиксом «spec».

## 1. Назначение и границы

**Делает** — по карточке, architecture 3.7 spec:
- держит семантику формата модуля — одна библиотека на всех потребителей. Вердикт и проверка ключей при публикации зависят от одного кода видов проверки (R17 spec). При обновлениях R03 spec держат тест-векторы формата — `format/vectors/` `spec`: CI `spec` проходит их своей второй реализацией, а CI `module-kit` — все (R42 spec; repositories 3 spec);
- читает модуль автора — YAML 1.2.2 с ограничениями формата — и проверяет его по JSON Schema 2020-12 и правилам формата. Отказ называет ID правила (R45 spec). Что проверяется, перечисляет карточка:
  - схема, граф навыков без циклов, ссылки, навыки-опоры общей таксономии и зависимости паспорта (R49 spec, R50 spec);
  - задачи, проверки ячеек, каталоги ловушек и приёмов, уроки (R51 spec, R56 spec, R57 spec, R58 spec);
  - шаблоны со всеми экземплярами и их решателями (R60 spec, R61 spec);
  - тексты автора, которые ученик видит до решения, — проверкой утечки (concepts 10 spec; R57 spec).

  О картинке с прозрачными пикселями CLI предупреждает, но не отказывает (вторая поправка 2026-11-01 к R51 spec);
- читает версии формата модуля в окне. Пока формат v0, в окне только текущая версия, 0.2. Модуль проверяется, собирается и исполняется по версии, которую называет его паспорт, а модуль вне окна отклоняется с причиной (R44 spec);
- собирает пакет модуля и индекс модулей (R17 spec):
  - пакет — канонический вид с ярусом каждого навыка и числами каждого экземпляра шаблона, часть для устройства и опись (R62 spec);
  - индекс — из индекса прошлого релиза, с записью каждой опубликованной версии по дайджесту пакета и с отметками о снятии (R63 spec).

  Пакет — подписанная версия модуля, адресуемая по содержимому, индекс — подписанный накопительный список опубликованных версий (concepts 15 spec). На релизе `modules` их собирает CLI `module-kit` (repositories 4 spec);
- сравнивает новую версию модуля с прошлой по правилам semver и сверяет упразднённые навыки (concepts 3 spec; R49 spec);
- заполняет шаблоны задач (concepts 9.3 spec; R60 spec):
  - перебирает сочетания параметров — не больше 10 000 на шаблон — и строит экземпляры;
  - ключ каждого экземпляра считает решатель ещё при проверке модуля (R61 spec);
  - экземпляр попытки выбирает генератор от её зерна (R59 spec);
- строит открытый вид задачи — без ключа, решения и ловушек — по пометкам `x-open` в схемах формата: поле без пометки закрыто (architecture 1 spec; R05 spec; R45 spec). Открытые виды он строит на сервере из части пакета для сервера, и в пакет они не входят (R62 spec);
- держит виды проверки и нормализацию: в формате v0 это `number`, `fraction`, `choice`, `predicate` и `rubric` (concepts 9.2 spec; R56 spec). Это код платформы, общий для всех модулей: модуль только выбирает вид и задаёт параметры (R02 spec);
- выносит вердикт: сравнивает наблюдение с ключом или условием (architecture 1 spec; concepts 9.1 spec; R03 spec). Вердикт допускается, только если ключ или условие дал автор, решатель или взрослый, а при ключе, выведенном ИИ, вердикта нет (concepts 9.4 spec; R04 spec);
- держит проверку утечки — поиск ключа и фраз-утечек, которые указал автор (concepts 10 spec; R57 spec). `api` проверяет ею текст наставника до того, как он уйдёт на устройство (architecture 3.5 spec), а проверка модуля — тексты автора (concepts 10 spec);
- исполняет Starlark в песочнице, как в standalone (repositories 7 spec):
  - условия проверки — не больше 1 000 000 шагов, а сработавший лимит или ошибка дают «без проверки» (R56 spec);
  - решатели шаблонов — не больше 100 000 000 шагов на все экземпляры шаблона (R61 spec).

  Часы и одновременные запуски — раздел 2.3;
- сверяет пакет по цепочке дайджестов от индекса, который закрепил pull request `api`: индекс → опись → части → файлы (R62 spec, R63 spec). Подписей библиотека не проверяет: подпись файла релиза проверяет рецепт `spec` `release-verify` там, где файл входит в NovaTrail (R77 spec; поправка 2026-11-01 к R17 spec);
- даёт автору CLI: проверить модуль, собрать пакет, построить индекс. Индекс CLI строит внутри общего релиза `modules`, после номера и до публикации (architecture 3.7 spec; R63 spec).

**Не делает:**
- не обращается к ИИ (architecture 3.7 spec). Работу читает `tutor`, а `module-kit` получает от `api` уже наблюдение (architecture 1 spec; architecture 3.5 spec);
- не ходит в сеть: библиотека — вычисления без ввода-вывода, а CLI только читает и пишет файлы модуля (architecture 3.7 spec; repositories 1 spec). Прошлый индекс сборка получает файлом: его скачивает и проверяет общий релиз `.github` (поправка 2026-11-02 к R63 spec);
- не хранит данных учеников (architecture 3.7 spec): из частей платформы их на сервере хранит только `api` (architecture 1 spec; R15 spec);
- не владеет форматом. Схемы, текст формата, примеры и тест-векторы лежат в `format/` `spec` (R42 spec; R45 spec; repositories 5 spec), а `module-kit` держит семантику формата в коде (R17 spec). CI `spec` проверяет примеры валидатором JSON Schema, а векторы — своей второй реализацией, без `module-kit`: иначе `spec` зависел бы от того, кто зависит от него (repositories 3 spec);
- не задаёт генератор случайности и вывод зёрен. Их задаёт `spec`, с тест-векторами, а `module-kit` их проходит (architecture 1 spec). Генератор — PCG из `math/rand/v2` Go, а места случайности выводятся, как у стенда standalone (R59 spec). Они одни на всю платформу — на `module-kit`, виджеты, `engine` и `sim` (R27 spec);
- не подписывает пакеты и индекс и не проверяет подписей: подпись ставит общий workflow релиза из `.github`, а проверяет рецепт `spec` `release-verify` (R17 spec; R77 spec; repositories 4 spec);
- не выбирает урок или задачу и не считает модель ученика: это `engine` и `learner-model` (architecture 1 spec). Типы графа у них простые и не из `module-kit`: задаёт их `learner-model`, а `engine` берёт у него (architecture 3.8 spec; R28 spec);
- не публикует модули: реестр v0 — подписанный индекс, который собирает CI `modules`. Сервис реестра появится на этапе 4 и тоже будет построен на `module-kit` (R17 spec).

**Данные и связи:**
- своих данных нет: он работает с модулем, пакетом и индексом, которые ему дают. Данных учеников и секретов нет, как ни в одном репозитории (architecture 3.7 spec; RULES.md, Hard rules). В процессе `api` он получает наблюдение и ключ или условие задачи и возвращает вердикт (architecture 3.5 spec);
- при сборке ни от чего не зависит, кроме `spec` (repositories 2.1 spec; repositories 3 spec):
  - схемы, примеры и векторы формата он берёт модулем Go `github.com/altnovatrail/spec` — `spec.Format` по тегу релиза, тому же, что у подмодуля `spec/`, — а не из подмодуля;
  - типы частей, их открытых видов, канонического вида, описи и индекса — из пакетов `types/formatv0` и `types/formatv0/open` того же модуля (R64 spec, R72 spec).

  Версию модуля `spec` в `go.mod` и подмодуль двигают вместе, переходом на тег, а из Dependabot `gomod` модуль `spec` исключают (R64 spec). Придёт это с первым кодом, который читает формат: пока модуля `spec` в `go.mod` нет (Д46 в таблице 4 RUN.md `spec`; T05);
- его импортируют `api` и eval-раннер `tutor`, а CLI по точному тегу берут CI `modules` и CI `canvas` — бинарником релиза, который сверяет `release-verify` до первого запуска (repositories 1 spec; repositories 3 spec; R77 spec). CI `canvas` берёт CLI, чтобы собрать страницу образцов: открытые виды и медиафайлы она получает при сборке, из пакетов релиза `modules` (R25 spec; architecture 3.2 spec). Нужна ли для этого своя команда CLI — раздел 2.5. Позже, на этапе 4, на нём строятся `module-registry` и `module-pipeline` (architecture 3.7 spec; repositories 2.1 spec);
- `sim` его не берёт: часть пакета для сервера он читает по типам канонического вида из `spec` (architecture 3.9 spec; R62 spec). `learner-model` и `engine` пакетов не читают: граф им передаёт `api` (architecture 3.8 spec);
- вердикт воспроизводим. Попытка хранит дайджест своего пакета — дайджест описи, — версию формата модуля, версию `module-kit` и зерно. Вердикт повторяется по сохранённому наблюдению, без ИИ, тем релизом `module-kit`, который его вынес (architecture 1 spec; R03 spec; R44 spec). Изменение поведения вида проверки — новая версия формата (R44 spec);
- новую модель ИИ допускают по eval, и ложные тревоги там считаются по вердикту, через `module-kit` (architecture 4.6 spec; repositories 4 spec);
- проверки CI пришли ещё до кода: общие проверки из `.github`, среди них `docs-check-repo`. Какие они и что держит мерж — R30 spec и architecture 3.11 spec. Подключила их T17 `spec`, и файлы CI сюда писала её сессия; файлы CI под свой код пишет сессия этого репозитория (поправка 2026-10-21 к R30 spec).
  - Находки CodeQL, Scorecard и SonarCloud закрываются перед новой задачей (правило 12 [RUN.md](../RUN.md)), а общие для всех репозиториев причины тех, что остаются открытыми, записаны в R33 spec.
  - Проверки Go пришли с заготовкой CLI (R04): рецепты у них свои, а запускает их обёртка Go — общий `go.yml` `.github`, который на ней и проверен (R16 .github). Go CodeQL читает с T12: общий `codeql.yml` `.github` v0.1.3 собирает его с Go из стадии `toolchain` (R06). Обёртка TypeScript проверяется на первом коде на TypeScript, у `canvas` (repositories 6 spec; architecture 3.11 spec).
  - Релиз выпускает общий `release.yml` `.github`, а файлы для него собирает свой рецепт `release-files` (R07; раздел 3.1).
  - Код приходит со своими тестами, а когда покрытие пойдёт в Codecov, мерж будет ждать `codecov/patch` (RULES.md, How we work; R48 spec) — Д8 в таблице 4 [RUN.md](../RUN.md);
- этап 1, вместе с `modules`: формат модуля сразу проверяется на трёх модулях двух предметов (repositories 6 spec; R08 spec; R41 spec; product 11 spec). Работу по порядку repositories 6 spec ведёт сквозной план — COORDINATION.md `spec`; у `module-kit` там строки 3.1–3.3: переход на линию 0.6, детализация и эпик кода (R76 spec).

## 2. Решить при детализации

Пункты «Решить при детализации» из карточки — по подразделу на пункт. Каждый решает своя задача эпика E1 в [RUN.md](../RUN.md). Решённый пункт называет запись журнала, а само решение переезжает в раздел 3. Вопрос, который выходит за этот репозиторий, уходит «в реестр spec» (RULES.md, Across repositories).

У карточки три пункта — подразделы 2.1–2.3. С v0.6.3 все три — не вопросы о формате, а то, что формат оставил `module-kit`: чем писать канонический текст пакета, как CI проходит векторы генератора, часы и одновременные запуски Starlark (architecture 3.7 spec; R08). Подраздел 2.4 — сверх карточки. Сравнение версий по semver карточка называет среди того, что `module-kit` делает (architecture 3.7 spec), а формат задаёт для него только мажорную версию и таблицу соответствия (concepts 3 spec; R49 spec). Этот пункт называет задача, которая завела репозиторий (T05 `spec`). Подраздел 2.5 — тоже сверх карточки: его вопрос `spec` по-прежнему оставляет детализации `module-kit` (поправка 2026-10-28 к R25 spec). Карточку под пункты сверх неё правит итог детализации, T05, — строкой таблицы 3 «в реестр spec», когда решение уже принято (architecture 3 spec).

На v0.2.13 два первых пункта были вопросами реестра `spec`, Т-15 и Т-18, и T01 и T02 должны были принести к ним варианты. Эпик формата E2 `spec` их не дождался: Т-15 закрыл R62 spec, Т-18 — R59 spec, и обе задачи сужены (R08).

### 2.1. Пакет модуля и канонический текст

- **Суть:** как `module-kit` собирает пакет модуля по R62 spec и чем пишет его канонический текст. Состав пакета решил формат, а здесь решается, как его исполнить.
- **Что известно:**
  - пакет — три части (R62 spec; [текст формата](../spec/format/README.md), раздел 25; concepts 15 spec):
    - часть для сервера — канонический вид, `server.json`: модуль одним JSON-объектом, части — как написаны, без значений по умолчанию, и сверх них `skill_tiers` и `instances`;
    - часть для устройства — `LICENSE`, `NOTICE` и медиафайлы, а когда в формате будут виджеты, — их бандлы. Список — `device.json`;
    - опись — `inventory.json`: версия формата, ID и версия модуля и обе части по дайджестам и размерам. Это не паспорт модуля, хотя паспорт в коде и схемах называется `manifest`;
  - канонический текст — RFC 8785: ключи по порядку кодовых единиц, без пробелов, UTF-8. Экранируются только `"`, `\` и управляющие знаки, а `/`, `<`, `>`, `&`, U+2028 и U+2029 пишутся как есть (там же; concepts 15 spec);
  - дайджест — `sha256:` и 64 строчные шестнадцатеричные цифры от канонического текста. Дайджест пакета — дайджест описи: его хранит попытка (concepts 15 spec; architecture 1 spec);
  - пакет зависит только от файлов модуля и версии формата его паспорта: версии `module-kit`, времени и путей в нём нет, а eval-набор в него не входит (R62 spec);
  - чем писать канонический текст, карточка оставила `module-kit`. `encoding/json` Go экранирует то, что текст формата пишет как есть, а `jsontext.Value.Canonicalize` в go1.27.3 — только под экспериментом `jsonv2` (architecture 3.7 spec). Рецепт `release-files` собирает с пустым `GOEXPERIMENT` (R07);
  - векторы — `format/vectors/tiers/` и `format/vectors/packages/` `spec`; результаты решателя вектора `with-solver` CI `module-kit` считает своим запуском (Д44 в таблице 4 RUN.md `spec`);
  - кто читает пакет:
    - через `module-kit` — `api`, по дайджесту, и eval-раннер `tutor`; с этапа 4 — `module-registry`, построенный на `module-kit` (architecture 3.10 spec; R17 spec);
    - через CLI `module-kit` — CI `canvas`, для страницы образцов (R25 spec);
    - без `module-kit` — `sim`, по типам канонического вида из `spec` (architecture 3.9 spec; R62 spec);
    - часть для устройства раздаёт `api` (architecture 3.10 spec);
  - на этапе 1 пакеты и индекс лежат в релизах `modules`, на этапе 2 пакеты лежат в образе `api`. Для повтора вердикта старые версии пакетов берутся из релизов `modules` (architecture 3.10 spec; Т-17).
- **Варианты:** пока пусто — их приносит задача.
- **Задача:** T01.

### 2.2. Генератор в CI

- **Суть:** чем `module-kit` исполняет генератор и места случайности по R59 spec — стандартной библиотекой Go или своим кодом — и как его CI проходит векторы `format/vectors/random/` `spec`.
- **Что известно:**
  - генератор — PCG со 128 битами состояния, ровно `PCG` пакета `math/rand/v2` Go, с выходом DXSM. Функции поверх него — `integer`, `below`, `between`, `fraction` и `permutation` (R59 spec; [текст формата](../spec/format/README.md), раздел 22);
  - место случайности — зерно и список имён. Его поток — генератор с двумя словами: FNV-1a 64 от зерна и имён через `|` и само зерно. Так новое место не сдвигает чисел прежних. Это схема стенда standalone, и его `seeded` — частный случай места (R59 spec; раздел 4);
  - текст формата задаёт константы числами, чтобы ни одна реализация не зависела от библиотеки, а `spec` считает векторы сам, второй реализацией (R59 spec);
  - экземпляр шаблона попытки выбирает одно число `below(count)` места `template`, ID задачи, `instance` (R60 spec);
  - генератора кодом модуль Go `spec` не даёт, и как звать его функции, решает `module-kit` (R72 spec);
  - тот же генератор берут `engine`, `sim` и виджеты (R27 spec; R59 spec). В песочнице решателей standalone генератора нет вовсе, а вывод зёрен по назначению был только в стенде `tools/learners` (раздел 4).
- **Варианты:** пока пусто.
- **Задача:** T02.

### 2.3. Часы, одновременные запуски и память Starlark

- **Суть:** сколько времени у вычисления условия проверки и у решателя шаблона, сколько их идёт сразу, кто держит часы и места, что значат сработавшие часы, сколько памяти у песочницы и нужен ли отдельный процесс. Решается здесь: шаги, диалект и место решателя решил формат, а часы и одновременные запуски он оставил `module-kit` (architecture 3.7 spec).
- **Что решил формат:**
  - условие проверки — выражение Starlark от целого `value`, не длиннее 1000 знаков и не больше 1 000 000 шагов в счёте go.starlark.net. Сработавший лимит, ошибка и часы `module-kit` дают «без проверки» (R56 spec; [текст формата](../spec/format/README.md), раздел 15);
  - выражения шаблона лимита шагов не требуют: их шагов не больше, чем знаков. Сочетаний — не больше 10 000 на шаблон (R60 spec);
  - решатель шаблона идёт только при проверке модуля, на каждом экземпляре по порядку. Программа — не длиннее 8192 знаков, бюджет — 100 000 000 шагов на все экземпляры шаблона вместе. Диалект — go.starlark.net с `Set`, `While`, `TopLevelControl` и `GlobalReassign`, без рекурсии и `load`, только с встроенными именами Starlark (R61 spec; [текст формата](../spec/format/README.md), раздел 24);
  - ключ шаблона считается заранее и лежит в `instances` пакета, а сервер выражений не считает и решателей не запускает (R60 spec; R62 spec);
  - смена этих лимитов и диалекта — новая версия формата (R44 spec).
- **Что известно из standalone** (раздел 4):
  - 25 000 000 шагов и 2 секунды на запуск, по одному запуску на vCPU (SPEC 6.6 standalone);
  - исходник решателя — не больше 8 КБ; рекурсия выключена: рекурсивная функция Starlark идёт по стеку Go, а переполнение стека уронило бы весь экземпляр, а не один запрос;
  - встроенные функции платят шаг за каждые 16 байт того, что строят, и не строят списков длиннее 1 000 000 элементов: интерпретатор считает шаги, но не память. Так шаги ограничивают и память, которую держат встроенные функции, — 381 МиБ на потолке шагов; среде Go отдана девять десятых памяти экземпляра (`GOMEMLIMIT`);
  - запуск, который 3 секунды не находит свободного места, не начинается, а задача считается непроверенной;
  - у памяти осталась известная дыра, и она шире одного оператора. Оператор `*` и повторное `xs = xs + xs` строят огромный список мимо счётчиков; литерал `{}` в цикле держит по 512 байт на каждые девять шагов, и экземпляр с 1 ГиБ памяти был убит именно так; методы самих значений — `x.extend(range(3000000))` — берут три миллиона элементов за девятнадцать шагов (SPEC 6.6 standalone; R53 standalone). Внутри одного вызова не видны и часы. Закрыло бы всё это только исполнение решателя в отдельном процессе с лимитом памяти, а v1 standalone на это не пошёл;
  - цену в байтах, один запуск на vCPU и 1 ГиБ памяти на каждый vCPU решил R125 standalone. Отвергнуты там четыре места с часами как страховкой — вердикт зависел бы от нагрузки — и отдельный процесс сразу;
  - лимит шагов калибровали по 450 эталонным решателям: худший потребовал 2 081 362 шага (R53 standalone); худший из 603 — 2 413 617 шагов;
  - интерпретатор — `go.starlark.net`, закреплённый псевдоверсией коммита.
- **Что осталось:**
  - часы условия входят в вердикт: сработав, они дают «без проверки», а это зависит от машины. Вердикт же повторяется по сохранённому наблюдению — тем релизом `module-kit`, который его вынес (architecture 1 spec; R03 spec; R44 spec);
  - часы решателя, который идёт только при проверке модуля — в CI `modules`, а с этапа 4 в `module-registry`: если они срабатывают раньше бюджета шагов, приём шаблона у границы бюджета зависит от машины;
  - кто держит часы и считает одновременные запуски — `module-kit` или `api`: библиотека — вычисления без ввода-вывода (repositories 1 spec);
  - память — о ней формат молчит. С этапа 4 модули публикуют и сторонние авторы, а `module-registry` построен на `module-kit` (R17 spec; architecture 3.13 spec): хватит ли лимитов внутри процесса или нужен отдельный процесс (О-9);
  - опции диалекта у условия: R61 spec задаёт их решателю, а условию — нет. Это поведение вида проверки, то есть формат;
  - версия go.starlark.net: счёт шагов от версии может меняться, а повтор вердикта держит версия `module-kit` в попытке (R44 spec).
- **Варианты:** пока пусто.
- **Задача:** T03.

### 2.4. Правила semver для `diff`

- **Суть:** по каким правилам `module-kit` сравнивает новую версию модуля с прошлой: что поднимает мажорную, минорную и патч-версию, что сравнение находит само, а что объявляет автор. Карточка называет сравнение и сверку упразднённых навыков в том, что `module-kit` делает, но не среди пунктов «Решить при детализации» (architecture 3.7 spec); его называет задача, которая завела этот репозиторий (T05 `spec`). Команды `diff` в карточке нет: CLI там — «проверить, собрать, построить индекс». Отдельная это команда или часть проверки, решает задача.
- **Что известно:**
  - версия модуля — semver `X.Y.Z`, без пред-релизов (R46 spec; concepts 3 spec);
  - мажорную поднимает несовместимое изменение: удаление навыка или изменение его смысла, изменение ключа опубликованной задачи (concepts 3 spec). Что поднимает минорную и патч-версию, `spec` не говорит;
  - таблица соответствия — `retired_skills` графа навыков: упразднённый навык и навыки, которые его заменили. Запись остаётся во всех следующих версиях, а ID упразднённого навыка не называет другого навыка (R49 spec; [текст формата](../spec/format/README.md), раздел 12);
  - смену смысла по модулю не видно, и автор объявляет её новым ID (R49 spec);
  - что каждый навык прошлой версии есть в этой или в `retired_skills` и ни одна запись не пропала, видно только сравнением двух версий. Это сравнение — `module-kit`, и его правила придут в текст формата строкой «в реестр spec» (R49 spec; Д30 в таблице 4 RUN.md `spec`);
  - переход модуля на новую версию формата — тоже новая версия модуля, а какую часть версии он поднимает, решают правила semver `module-kit` (R44 spec). Так же — правка `teaches` и `dependencies` паспорта (R50 spec; Д31 там же);
  - по таблице соответствия переносят прогресс `api` и `learner-model` после мажорной версии модуля (architecture 3.5, 3.8 spec);
  - старые версии пакетов берутся из релизов `modules` (architecture 3.10 spec). Откуда CI `modules` берёт прошлую версию модуля для сравнения, решает T01 modules;
  - в каноническом виде части лежат как написаны, без значений по умолчанию (R62 spec): смену поведения при переходе формата там видно только по `module_format_version`;
  - версия модуля — не версия формата. Версию формата — MAJOR.MINOR — называет паспорт, полем `module_format_version` (R44 spec; R46 spec). Здесь речь только о версии модуля.
- **Варианты:** пока пусто.
- **Задача:** T04.

### 2.5. Команда CLI для открытых видов

- **Суть:** нужна ли CLI `module-kit` своя команда, которая отдаёт открытые виды задач сборке страницы образцов `canvas`, или этой сборке хватает команд карточки — проверить, собрать, построить индекс (architecture 3.7 spec). Вопрос R25 spec оставил детализации `module-kit`, и поправка 2026-10-28 к R25 spec оставила его там же. Решается здесь: CLI — часть `module-kit` (R17 spec). Новая команда меняет карточку, и её правку итог детализации отправляет строкой таблицы 3 «в реестр spec» (раздел 2).
- **Что известно:**
  - страница образцов — отдельная сборка `canvas` для образцов eval: образец делается той же канвой, что у учеников, и сохраняется файлом на устройстве (architecture 3.2 spec). К первому замеру она без хостинга: сборка скачивается из релиза `canvas` и открывается на устройстве, а задачи и медиафайлы кладутся в неё при сборке, из релиза `modules` через CLI `module-kit`. Во время работы страница по сети не ходит (R25 spec);
  - пакет `canvas` предметов не знает: данные получает только сборка страницы. CI `canvas` берёт CLI `module-kit` и релиз `modules` по точному тегу — это связь «инструмент CI» (R25 spec; repositories 1, 3 spec) — и сверяет оба рецептом `release-verify` (R77 spec);
  - CLI только читает и пишет файлы модуля и в сеть не ходит (architecture 3.7 spec): релиз `modules` скачивает сам CI `canvas`;
  - открытые виды CLI строит из части пакета для сервера, а в пакет они не входят; медиафайлы лежат в части для устройства (поправка 2026-10-28 к R25 spec; R62 spec);
  - открытый вид — по пометкам `x-open` в схемах формата (раздел 1; R45 spec). У шаблона до приложения доходит открытый вид экземпляра, а не сам шаблон (R60 spec): какие экземпляры отдаёт такая команда, решает она;
  - отдельная команда CLI решается и для сравнения версий — `diff` (раздел 2.4).
- **Варианты:** пока пусто.
- **Задача:** T05.

## 3. Устройство

Сюда переезжает то, что решено по разделу 2: из каких частей компонент состоит, его типы и данные, как он связан с другими репозиториями. Пока здесь только то, с чего начался код (R04).

### 3.1. Код, проверки и релизы

- **Модуль** — `github.com/altnovatrail/module-kit`. В `go.mod` — `go 1.27.0`, нижняя граница для тех, кто его импортирует. Сам Go 1.27.3 стоит в стадии `toolchain` devcontainer, где `GOTOOLCHAIN=local`: это Go самого `spec`, и контейнер `module-kit` от контейнера `spec` не отличается (R04; R08).
- **CLI** — `cmd/module-kit`. Пока он отвечает только на `--version`; проверка модуля, сборка пакета и индекса придут с кодом формата 0.2 (раздел 1).
- **Проверки** — рецепты `justfile`. Общий `go.yml` `.github` запускает их по одному в job, в стадии `toolchain` (R16 .github):

  | Рецепт | Что делает | Проверка в ruleset |
  |---|---|---|
  | `ci-lint` | gofmt по файлам `.go` модуля, без `testdata/` и `spec/`, и `go vet` | «Go / lint» |
  | `ci-test` | `go test -race -count=1 ./...` | «Go / test» |
  | `ci-vuln` | govulncheck v1.9.0 — известные уязвимости в коде, до которого модуль доходит | «Go / vuln» |
  | `ci-licenses` | go-licenses v2.0.1 — лицензии пакетов, из которых собирается модуль, по списку `ALLOWED_LICENSES` из `go.yml` | «Go / licenses» |

  Инструменты идут через `go run` по точной версии модуля, и их поднимают руками (R04). Go поднимается переходом на тег `spec`, где его поднял `spec` (R08). Своих исключений лицензий нет.
- **CodeQL** читает и Go: общий `codeql.yml` `.github` v0.1.3 со входом `languages: go` собирает код с Go из стадии `toolchain`, а проверка — «CodeQL / analyze (go)» (R06).
- **Релиз** — `.github/workflows/release.yml` зовёт общий `release.yml` `.github` v0.1.4 по SHA, после зелёного CI на `main` или по кнопке (R07; R18 .github):
  - номер — следующий патч линии `RELEASE_LINE` `0.1`, первый — v0.1.0; считают его рецепты релиза `spec`. Релиз выходит, только если с прошлого изменилось что-то из `paths` — всё, кроме `docs/`, файлов `*.md`, `.github/` и подмодуля `spec`. Его дают код и тесты, `go.mod`, `justfile`, `LICENSE` и весь `.devcontainer/`, даже если бинарник от правки не меняется; не дают — документы, CI, в том числе смена `RELEASE_LINE`, и тег `spec` (R07). Перечень вместо исключений, по шаблону architecture 3.3 .github, — Д7 в таблице 4 [RUN.md](../RUN.md);
  - файлы собирает рецепт `release-files <версия>` — в стадии `toolchain` выпускаемого коммита, из его дерева без git и подмодулей, в `dist/`, которую он создаёт с нуля:

    | Файл | Что это |
    |---|---|
    | `module-kit_<версия>_<os>_<arch>`, у windows — с `.exe` | CLI для linux, darwin и windows, на amd64 и arm64: `CGO_ENABLED=0`, `-trimpath`, `-buildvcs=false`, `-ldflags "-s -w -X main.version=<версия>"` |
    | `LICENSE` | Apache-2.0 самого `module-kit` |
    | `LICENSE-go` | лицензия Go, BSD-3-Clause: бинарник несёт его стандартную библиотеку |
    | `SHA256SUMS` | суммы sha256 всех файлов выше |

    Сборка воспроизводима: окружение Go вызывающего в неё не попадает, и из дерева `git archive` и из рабочей копии она даёт те же суммы. CLI, поставленный `go install` по тегу, берёт версию из сведений о сборке;
  - каждый файл общий workflow подписывает cosign без ключей, кладёт рядом его `.sigstore.json`, сверяет подписи до публикации, а после — `gh release verify` и `verify-asset`. Релизы неизменяемые, а теги `v*` держит ruleset 25559103 — условие проверки подписи (R77 spec);
  - подпись проверяет рецепт `spec` `release-verify`. Он пускает только подпись общего `release.yml` на релизе `.github` не ниже v0.1.7, поэтому релизы на v0.1.4 — v0.1.0 и тот, что дал переход на v0.6.3, — отклоняет. Закрепление до первого релиза с кодом поднимает Dependabot, T08 (R77 spec; R07; R08).
- **Чего ещё нет:** gopls в образе — Д5, покрытия в Codecov — Д8, проекта OpenSSF Best Practices — Д9 в таблице 4 [RUN.md](../RUN.md).

## 4. Источники

Внешний факт — условия поставщика, цена, закон, документация библиотеки — встаёт сюда с источником и датой прочтения (RULES.md, The bar).

- **Песочница решателей standalone** (разделы 2.2 и 2.3) — [mathtrail/mathtrail-standalone, `SPEC.md`](https://github.com/mathtrail/mathtrail-standalone/blob/c6cbeb2/SPEC.md), раздел 6 «The Starlark solver», ветка `main`, коммит c6cbeb2, прочитано 2026-10-19:
  - 6.1: «the program is written by the same model that wrote the task»;
  - 6.4, рекурсия выключена: «A recursive Starlark function recurses on the Go stack, and a stack overflow in Go cannot be recovered: it would take the whole instance down, not the request»;
  - 6.5: «Every helper that builds a list is capped at 1,000,000 elements», «a step for every sixteen bytes»; «There is no random number generator, seeded or otherwise. A brute force that samples proves nothing about uniqueness, and a verdict that depends on a seed is a verdict nobody can reproduce from the task alone»;
  - 6.6, таблица «The limits»: Steps — 25,000,000; Wall clock — 2 seconds per run; Memory — «What the built-ins keep: 381 MiB at the ceiling above; nine tenths of the instance for the runtime»; Source size — 8 KB; Result — 5 letters; Concurrent runs — 1 per vCPU; Wait for a slot — «3 seconds per run», и запуск, который не дождался места, «does not start, and the task is told it was not checked and cost no attempt». О литерале: «`[{} for i in range(n)]` keeps 512 bytes for every nine steps — 2.6 million dictionaries at the ceiling, and an instance of 1 GiB was killed for its memory by them». О калибровке: «The bench measured the worst of the 450 at 2,081,362 steps, which is what moved the ceiling from 10,000,000 to 25,000,000» (R53 standalone) и «the worst of all 603 is 2,413,617 steps». О дыре памяти: «a solver written that way takes the instance down rather than the request» и «The fix, if the gap ever bites, is a solver in a process of its own with a memory limit on it — a change of shape rather than a setting, and not one v1 pays for».
- **Дыра памяти шире, чем в SPEC** (раздел 2.3) — [`docs/decisions.md`](https://github.com/mathtrail/mathtrail-standalone/blob/c6cbeb2/docs/decisions.md) standalone, R53 standalone, коммит c6cbeb2, прочитано 2026-10-19: «the methods of the values themselves take a lazy range of any length in one step — `x.extend(range(3000000))` is nineteen steps», «so no cap on the helpers can close the gap, and only a limit on the memory of the process can»; о времени: «the clock is seen only between instructions» и «a process of its own would bound its time as well as its memory».
- **Цена в байтах, места и память экземпляра** (раздел 2.3) — R125 standalone, тот же `docs/decisions.md`, коммит c6cbeb2, прочитано 2026-10-19: «an instance runs one solver for each vCPU, and it has 1 GiB, nine tenths of it told to the runtime»; отвергнуто: «Four slots with the clock as insurance. Verdicts depend on the load» и «The process of its own now. A change of shape, and the author's decision»; открыто: «One account handing in solvers that spend their whole clock, for requests nobody opened, keeps an instance's single slot busy» и «Such a hand-in spends no attempt, so the account pays nothing for it. That stays open».
- **Интерпретатор Starlark standalone** (раздел 2.3) — [`go.mod`](https://github.com/mathtrail/mathtrail-standalone/blob/c6cbeb2/go.mod), коммит c6cbeb2, прочитано 2026-10-19: `go.starlark.net v0.0.0-20260908191801-89a6a09411d5`.
- **Случайность в коде standalone** (раздел 2.2) — поиск по `internal/` и `cmd/` того же коммита, 2026-10-19:
  - вне тестов `math/rand/v2` импортируют только `internal/store/drive/calls.go` и `internal/transport/mcp/afresh.go`, и оба вызова дают случайную длительность паузы (`time.Duration`); генератора задач от зерна там нет;
  - в тестах генераторы PCG получают зерно числами, без хеша: `rand.NewPCG(uint64(grade), uint64(child))` в `internal/domain/tutor/simulation_test.go`, `rand.NewPCG(7, 7)` в `pause_test.go`, `rand.NewPCG(d.Dice, 0)` в `property_test.go` и другие.
- **Вывод зёрен в стенде standalone** (раздел 2.2) — [`tools/learners/child.go`](https://github.com/mathtrail/mathtrail-standalone/blob/c6cbeb2/tools/learners/child.go), коммит c6cbeb2, прочитано 2026-10-19: функция `seeded(unit, purpose string) *rand.Rand` с комментарием «seeded is the random stream of one unit of the experiment for one purpose»; она пишет в `fnv.New64a()` строку `"%d|%s|%s|%s"` из `masterSeed`, `experiment`, `unit` и `purpose` и возвращает `rand.New(rand.NewPCG(hash.Sum64(), masterSeed))` из `math/rand/v2`. Та же функция есть в `research/experiments/` того же коммита.
- **Go 1.27.3** (R04) — список релизов go.dev, `https://go.dev/dl/?mode=json`, прочитано 2026-10-21: новейший стабильный релиз — `go1.27.3`; sha256 архива `go1.27.3.linux-amd64.tar.gz` — `65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40`, `go1.27.3.linux-arm64.tar.gz` — `a06e99309dbcbdca1a8ccd2282c3ee146e55930886e26b4f1193269587165f49`. Оба архива скачаны и сверены `sha256sum` 2026-10-21.
- **`GOTOOLCHAIN=local`** (R04) — [Go Toolchains](https://go.dev/doc/toolchain), прочитано 2026-10-21: «When GOTOOLCHAIN is set to local, the go command always runs the bundled Go toolchain».
- **Инструменты** (R04) — прокси модулей Go, `https://proxy.golang.org/<модуль>/@latest`, прочитано 2026-10-21: `golang.org/x/vuln` — v1.9.0 (2026-10-11), `github.com/google/go-licenses/v2` — v2.0.1 (2025-09-08). Лицензии — API GitHub, `repos/<владелец>/<репозиторий>/license`, 2026-10-21: golang/vuln — BSD-3-Clause, google/go-licenses — Apache-2.0.
- **`go-licenses check`** (R04) — [README v2.0.1](https://github.com/google/go-licenses/blob/v2.0.1/README.md), раздел Usages, Check, прочитано 2026-10-21: «Allow only specific license names» — `go-licenses check <package> [package...] --allowed_licenses=<comma separated license names>`. Проба 2026-10-21: со списком `spec` рецепт `ci-licenses` проходит, а со списком из одной MIT отвечает «Not allowed license 'Apache-2.0' found for library 'github.com/altnovatrail/module-kit/cmd/module-kit'».
- **Образец рецептов** (R04) — [`justfile` canvas-api](https://github.com/MathTrail/canvas-api/blob/0901af0439280334b3da33ffc4cc5c853e4403ff/justfile), коммит 0901af0, прочитано 2026-10-21: `ci-lint`, `ci-test`, `ci-vuln` и `ci-licenses`, а об инструментах — «Each is a Go program run straight from its module, so none of them is installed anywhere».
- **SonarCloud, docker:S8545** (R04) — правило «Go dependencies should be locked to verified versions» (API SonarCloud, `api/rules/show?key=docker:S8545`, 2026-10-21). На pull request #8 оно отметило строку `go install` gopls в Dockerfile: «Dependency versions are not predictable. Use a lock-file enforcing command instead». У canvas-api такие находки открыты у двух строк его Dockerfile, 198 и 199, хотя там стоит `-mod=readonly` (API SonarCloud, проект MathTrail_canvas-api, 2026-10-21).
- **Autobuild Go в CodeQL** (R06) — [GitHub Docs, «CodeQL build options and steps for compiled languages»](https://docs.github.com/en/code-security/reference/code-scanning/codeql/build-options-for-compiled-languages), раздел «Autodetection for Go», исходник в `github/docs` на коммите c008b66, прочитано 2026-10-21: «Invoke `make`, `ninja`, `./build` or `./build.sh` (in that order) until one of these commands succeeds and a subsequent `go list ./...` also succeeds»; «By default, test code (code in files ending in `_test.go`) is not analyzed».
- **Лицензии в релизе** (R07):
  - Apache-2.0 — [LICENSE](../LICENSE), раздел 4 (a): «You must give any other recipients of the Work or Derivative Works a copy of this License»;
  - лицензия Go — `LICENSE` в корне Go 1.27.3 стадии `toolchain`, `$(go env GOROOT)/LICENSE`, прочитано 2026-10-21: «Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution». Бинарник Go несёт стандартную библиотеку, а сторонних модулей у `module-kit` нет (`go.mod`).
- **Общий релиз** (R07) — [`release.yml`](https://github.com/altnovatrail/.github/blob/11435dbe61d3d4885c9a8dd30686e8111e26cabb/.github/workflows/release.yml) и R18 .github, релиз `.github` v0.1.4, коммит 11435db, прочитано 2026-10-21: условия рецепту `release-files` — стадия `toolchain` выпускаемого коммита, его дерево из `git archive`, запуск под uid раннера с `HOME=/tmp` и без токенов, в `dist/` — только простые файлы с именами по `^[A-Za-z0-9_+][A-Za-z0-9._+-]*$` и без `.sigstore.json`. Релиз неизменяемый: `gh release verify v0.1.4 -R altnovatrail/.github` — «Release v0.1.4 verified!».
