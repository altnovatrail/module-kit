# module-kit — архитектура

> Статус: каркас из заготовки `spec`, 2026-10-19; сверен с `spec` v0.2.11 2026-10-20 (R03) и с v0.2.13 2026-10-21 (R05); первый код — заготовка CLI, 2026-10-21 (R04); релизы — общим `release.yml` `.github`, с 2026-10-21 (R07).
>
> Что делает компонент, чего не делает и что о нём ещё предстоит решить. Это подробности его карточки — architecture 3.7 spec. Карточка остаётся сводкой: если подробности меняют то, что в ней сказано, она правится в `spec` тем же путём, что и вопросы в его реестры (RULES.md, Across repositories).
>
> Почему решено так, а не иначе, — в [decisions.md](decisions.md); план работ — в [RUN.md](../RUN.md). Понятия — concepts spec. Документы и решения `spec` лежат в `spec/docs/` и цитируются с суффиксом «spec».

## 1. Назначение и границы

**Делает:**
- держит семантику формата модуля — одна библиотека на всех потребителей. Вердикт и проверка ключей при публикации зависят от одного кода видов проверки, а одни тест-векторы держат R03 spec при обновлениях (R17 spec; Т-18);
- читает модуль автора и проверяет его (architecture 3.7 spec; R17 spec):
  - схему — JSON Schema модуля из `spec`, как рекомендует Т-7;
  - граф навыков без циклов (concepts 4 spec);
  - ссылки, в том числе на навыки общей таксономии: только через неё модуль ссылается на навыки вне себя (concepts 4 spec; Т-14);
  - ключи — решателем, запасные тексты — проверкой утечки (concepts 9.3, 10 spec);
- собирает пакет модуля и индекс модулей (R17 spec). Пакет — подписанная версия модуля, адресуемая по содержимому, индекс — подписанный список пакетов (concepts 15 spec). На релизе `modules` их собирает CLI `module-kit` (repositories 4 spec);
- сравнивает новую версию модуля с прошлой по правилам semver (architecture 3.7 spec; concepts 3 spec);
- заполняет шаблоны задач: числа подставляются от зерна, а ключ считает решатель (concepts 9.3 spec);
- строит открытый вид задачи — без ключа, решения и ловушек — по списку открытых полей: новое поле формата закрыто, пока его явно не открыли (architecture 1 spec; R05 spec);
- держит виды проверки и нормализацию. Это код платформы, общий для всех модулей: модуль только выбирает вид и задаёт параметры (concepts 9.2 spec; R02 spec);
- выносит вердикт: сравнивает наблюдение с ключом или условием (architecture 1 spec; concepts 9.1 spec; R03 spec). Вердикт допускается, только если ключ или условие дал автор, решатель или взрослый, а при ключе, выведенном ИИ, вердикта нет (concepts 9.4 spec; R04 spec);
- держит проверку утечки — поиск ключа и фраз-утечек, которые указал автор (concepts 10 spec). `api` проверяет ею текст наставника до того, как он уйдёт на устройство (architecture 3.5 spec), а проверка модуля — запасные тексты автора (concepts 10 spec);
- исполняет решатели на Starlark в песочнице, как в standalone (architecture 3.7 spec; repositories 7 spec);
- проверяет подпись пакета (architecture 3.7 spec; R17 spec). Что именно проверяется — Т-27. Его рекомендация — сверять и SAN — путь общего `release.yml` `.github` с SHA одного из релизов `.github`, — и репозиторий и workflow того, кто этот общий workflow вызвал. Условие такой проверки — неизменяемые релизы `.github` и ruleset его тегов `v*`: без них тег можно перенести на другой коммит. Решает его T37 `spec` по первому релизу `module-kit` (R07), и там же — откуда берётся список релизов `.github`: сама библиотека в сеть не ходит;
- даёт автору CLI: проверить модуль, собрать пакет, построить индекс (architecture 3.7 spec).

**Не делает:**
- не обращается к ИИ (architecture 3.7 spec). Работу читает `tutor`, а `module-kit` получает от `api` уже наблюдение (architecture 1 spec; architecture 3.5 spec);
- не ходит в сеть: библиотека — вычисления без ввода-вывода, а CLI только читает и пишет файлы модуля (architecture 3.7 spec; repositories 1 spec);
- не хранит данных учеников (architecture 3.7 spec): из частей платформы их на сервере хранит только `api` (architecture 1 spec; R15 spec);
- не владеет схемой: JSON Schema модуля, пакета и индекса лежит в `spec` (repositories 5 spec; Т-7, Т-15), а `module-kit` держит семантику формата (R17 spec). CI `spec` проверяет примеры модулей обычным валидатором JSON Schema, а не `module-kit`: иначе `spec` зависел бы от того, кто зависит от него (repositories 3 spec);
- не задаёт генератор случайности и вывод зёрен: по принципу воспроизводимости их задаёт `spec`, с тест-векторами, а `module-kit` их проходит (architecture 1 spec). Они одни на всю платформу — на `module-kit`, виджеты, `engine` и `sim` (R27 spec). Сам алгоритм — Т-18, он ещё открыт;
- не подписывает пакеты и индекс: подпись ставит общий workflow релиза из `.github` (R17 spec; repositories 4 spec);
- не выбирает урок или задачу и не считает модель ученика: это `engine` и `learner-model` (architecture 1 spec). Типы графа у них простые и не из `module-kit`: задаёт их `learner-model`, а `engine` берёт у него (architecture 3.8 spec; R28 spec);
- не публикует модули: реестр v0 — подписанный индекс, который собирает CI `modules`. Сервис реестра появится на этапе 4 и тоже будет построен на `module-kit` (R17 spec).

**Данные и связи:**
- своих данных нет: он работает с модулем, пакетом и индексом, которые ему дают. Данных учеников и секретов нет, как ни в одном репозитории (architecture 3.7 spec; RULES.md, Hard rules). В процессе `api` он получает наблюдение и ключ или условие задачи и возвращает вердикт (architecture 3.5 spec);
- при сборке ни от чего не зависит, кроме `spec` (repositories 2.1 spec; repositories 3 spec);
- его импортируют `api` и eval-раннер `tutor`, а CLI по точному тегу берут CI `modules` и CI `canvas` (repositories 1 spec; repositories 3 spec). CI `canvas` берёт CLI, чтобы собрать страницу образцов: задачи и медиафайлы она получает при сборке, из релиза `modules` (R25 spec; architecture 3.2 spec). Нужна ли для этого своя команда CLI — раздел 2.5. Позже, на этапе 4, на нём строятся `module-registry` и `module-pipeline` (architecture 3.7 spec; repositories 2.1 spec);
- `sim` его не берёт: пакеты модулей он читает по типам канонического вида из `spec` (architecture 3.9 spec; Т-15). `learner-model` и `engine` пакетов не читают: граф им передаёт `api` (architecture 3.8 spec);
- вердикт воспроизводим: попытка хранит дайджест своего пакета, версию формата и зерно, а вердикт повторяется по сохранённому наблюдению, без ИИ (architecture 1 spec; R03 spec). По рекомендации Т-18 изменение поведения вида проверки поднимает версию формата;
- новую модель ИИ допускают по eval, и ложные тревоги там считаются по вердикту, через `module-kit` (architecture 4.6 spec; repositories 4 spec);
- проверки CI пришли ещё до кода: общие проверки из `.github`, среди них `docs-check-repo`. Какие они и что держит мерж — R30 spec и architecture 3.11 spec. Подключила их T17 `spec`, и файлы CI сюда писала её сессия, а не задача этого репозитория. Находки CodeQL, Scorecard и SonarCloud закрываются перед новой задачей (правило 12 [RUN.md](../RUN.md)), а общие для всех репозиториев причины тех, что остаются открытыми, записаны в R33 spec. Проверки Go пришли с заготовкой CLI (R04): рецепты у них свои, а запускает их обёртка Go — общий `go.yml` `.github`, который на ней и проверен (R16 .github). Релиз выпускает общий `release.yml` `.github`, а файлы для него собирает свой рецепт `release-files` (R07; раздел 3.1). Обёртка TypeScript проверяется на первом коде на TypeScript, у `canvas` (repositories 6 spec; architecture 3.11 spec). Go CodeQL читает с T12: общий `codeql.yml` `.github` v0.1.3 собирает его с Go из стадии `toolchain` (R06);
- этап 1, вместе с `modules`: формат модуля v0 сразу проверяется на двух предметах (repositories 6 spec; R08 spec; product 11 spec).

## 2. Решить при детализации

Пункты «Решить при детализации» из карточки — по подразделу на пункт. Каждый решает своя задача эпика E1 в [RUN.md](../RUN.md). Решённый пункт называет запись журнала, а само решение переезжает в раздел 3. Вопрос, который выходит за этот репозиторий, уходит «в реестр spec» (RULES.md, Across repositories).

У карточки три пункта — подразделы 2.1–2.3. Подраздел 2.4 — сверх карточки: сравнение версий по semver карточка называет среди того, что `module-kit` делает (architecture 3.7 spec), но правила для него `spec` задаёт только для мажорной версии (concepts 3 spec). Этот пункт называет задача, которая завела репозиторий (T05 `spec`). Подраздел 2.5 — тоже сверх карточки: его вопрос R25 spec оставил детализации `module-kit`. Карточку под пункты сверх неё правит итог детализации, T05, — строкой таблицы 3 «в реестр spec», когда решение уже принято (architecture 3 spec).

Два первых пункта — вопросы реестра `spec`, Т-15 и Т-18. Их закрывает эпик формата модуля v0 — E2 в RUN.md `spec`, вместе со схемами пакета и индекса и с тест-векторами (repositories 5 spec). Здесь задачи приносят варианты и рекомендацию от того, кто будет это исполнять, а то, что закрывает вопрос, уходит «в реестр spec». Задачи этих двух пунктов, T01 и T02, идут одновременно с E2, а не до него: так `spec` ответил на В19. С v0.2.12 E2 — текущий эпик `spec`: Т-15 закрывает его T34, которая ждёт T01, а Т-18 — T32, которая ждёт T02 (таблица 1 RUN.md `spec`; R05).

### 2.1. Что внутри пакета (Т-15)

- **Суть:** из чего состоит пакет модуля, который собирает `module-kit`: YAML автора как есть или скомпилированный канонический вид, какие у пакета части, как считается его дайджест и что из него читает каждый потребитель. От этого зависят и тест-векторы (раздел 2.2), и то, что сравнивает semver (раздел 2.4). Вопрос — реестра `spec` (Т-15): схему пакета и индекса пишет `spec` (repositories 5 spec).
- **Что известно:**
  - варианты Т-15 — YAML автора как есть или скомпилированный канонический вид. Рекомендация Т-15 — скомпилированный: `module-kit` собирает YAML в канонический вид по схеме из `spec`, а `api`, `sim` и eval читают только его;
  - части пакета, по рекомендации Т-15: часть для сервера — всё содержание, в том числе ключи, решения и ловушки; часть для устройства — только медиафайлы и бандлы виджетов, без ключей. У каждой свой дайджест, а дайджест пакета — дайджест того, что называет дайджесты обеих частей (architecture 3.10 spec). `spec` пока зовёт это манифестом пакета;
  - «манифест пакета» — не паспорт модуля, хотя паспорт в коде и схемах называется `manifest` (concepts 15 spec): одно слово для двух предметов. `spec` решил так: свой термин, без слова «манифест», то, что называет дайджесты обеих частей, получит в concepts 15 spec вместе со схемой пакета, в эпике E2 (Т-15; architecture 3.10 spec; ответ на В8 в RUN.md `spec`). Варианты термина приносит T01;
  - пакет модуля — подписанная версия модуля, адресуемая по содержимому, индекс модулей — подписанный список пакетов (concepts 15 spec). Подпись ставит общий workflow релиза из `.github`, а `module-kit` её проверяет (R17 spec); что проверка сверяет — Т-27 (раздел 1);
  - кто читает пакет:
    - через `module-kit` — `api`, по дайджесту из своей конфигурации, с этапа 2, и eval-раннер `tutor` (architecture 3.10 spec; repositories 2.1 spec; repositories 4 spec); с этапа 4 — `module-registry`, построенный на `module-kit` (R17 spec);
    - через CLI `module-kit` — CI `canvas`: сборка страницы образцов берёт задачи и медиафайлы из релиза `modules` (R25 spec; architecture 3.2 spec). Нужна ли ей своя команда — раздел 2.5;
    - без `module-kit` — `sim`, по типам канонического вида из `spec` (architecture 3.9 spec). Тогда эти типы — часть `spec`, а не этого репозитория;
    - часть для устройства получает `app`: медиафайлы ему раздаёт `api` (architecture 3.10 spec), а бандлы виджетов с этапа 4 открывает хост виджетов из `canvas` (architecture 3.1 spec; R16 spec);
  - на этапе 1 пакеты и индекс лежат в релизах `modules`, на этапе 2 пакеты лежат в образе `api`, а медиафайлы раздаёт `api`. Для повтора вердикта старые версии пакетов берутся из релизов `modules` (architecture 3.10 spec; Т-17);
  - попытка хранит дайджест своего пакета (architecture 1 spec). Даёт ли один и тот же модуль при повторной сборке тот же дайджест — то есть воспроизводима ли сборка пакета, — `spec` пока не говорит;
  - открытый вид строит `module-kit` на сервере, по списку открытых полей (architecture 1 spec; R05 spec), а в части для устройства, по рекомендации Т-15, — только медиафайлы и бандлы виджетов;
  - бандл виджета закреплён по хешу в пакете модуля (R16 spec), но виджетов в первом срезе нет (product 10 spec; architecture 3.4 spec).
- **Варианты:** пока пусто — их приносит задача.
- **Задача:** T01.

### 2.2. Генератор случайности и тест-векторы (Т-18)

- **Суть:** каким алгоритмом шаблон задачи получает числа от зерна, как из зерна выводятся зёрна для разных нужд и какие тест-векторы держат одно и то же поведение у всех, кто их проходит. Вопрос — реестра `spec` (Т-18): генератор случайности и вывод зёрен с тест-векторами — нормативная часть `spec` (repositories 5 spec). `module-kit` исполняет их первым, поэтому варианты приносит его задача.
- **Что известно:**
  - варианты Т-18 — генератор языка программирования или свой алгоритм, заданный в `spec`. Рекомендация Т-18 — алгоритм и вывод зёрен задаёт `spec`, с тест-векторами: фрагмент модуля и зерно дают точный открытый вид, а наблюдение и ключ — точный вердикт. `module-kit`, виджеты, `engine` и `sim` берут один генератор и одну схему зёрен и проходят тест-векторы в CI; алгоритм выбирается с учётом схемы стенда standalone — FNV-64a и PCG (R27 spec). Изменение поведения вида проверки поднимает версию формата, а попытка хранит её вместе с дайджестом пакета и зерном. Принцип воспроизводимости в architecture 1 spec говорит то же;
  - зерно нужно не только шаблонам. Случайность движка — только от зерна, и выбор движка хранит зерно (concepts 12 spec; architecture 1 spec); случайность виджета — тоже только от зерна (concepts 8 spec). Поэтому вывод зёрен — общий. R27 spec так и решил: генератор и вывод зёрен из `spec` — один на `module-kit`, виджеты, `engine` и `sim`, а алгоритм выбирает эпик E2 в RUN.md `spec` с рекомендациями двух задач — T02 этого репозитория и T03 engine;
  - виджет — код автора модуля, он работает на устройстве в iframe-песочнице (R16 spec), а `module-kit` написан на Go (R11 spec). По рекомендации Т-18 виджеты проходят те же тест-векторы, а по рекомендации Т-24 сервер повторяет состояние виджета по зерну и событиям ученика. Тогда алгоритм должен давать одни и те же числа в реализациях на разных языках. Виджеты — этап 4 (architecture 3.4 spec);
  - ключ шаблонной задачи считает решатель, и происхождение такого ключа — «проверено решателем»: при нём вердикт возможен (concepts 9.3, 9.4 spec);
  - генерация задач на лету — открытый вопрос О-12. Его рекомендация — пул и шаблоны с параметрами: числа — от зерна, ключ — от решателя, а генерация ИИ — только при подготовке модуля;
  - в standalone задачи писала модель, а решатель их проверял (О-12; SPEC 6.1 standalone). В его песочнице генератора случайности нет вовсе, даже с зерном: «a verdict that depends on a seed is a verdict nobody can reproduce from the task alone» (SPEC 6.5 standalone). Генератора задач от зерна в standalone нет (раздел 4);
  - зато вывод зёрен по назначению в standalone есть — в стенде виртуальных учеников `tools/learners`, который переносится в `sim` (repositories 7 spec). Каждый поток случайности там — свой генератор PCG из `math/rand/v2` с двумя словами зерна: хешем FNV-64a от строки «главное зерно | эксперимент | единица | назначение» и самим главным зерном (раздел 4). Так новое назначение не сдвигает числа в старых. В тестах `internal/` standalone генераторы PCG получают зерно прямо числами, без хеша, — например, класс и номер ребёнка; одной схемы на весь standalone нет (раздел 4). Двух схем на платформе не будет: `sim` выводит зёрна генератором из `spec`, а не своей схемой (R27 spec; architecture 3.9 spec), а алгоритм `spec` выбирается с учётом схемы стенда (Т-18);
  - версия формата — не версия модуля: версию формата называет паспорт (concepts 3 spec), а правила версий `spec` — Т-11.
- **Варианты:** пока пусто.
- **Задача:** T02.

### 2.3. Лимиты песочницы Starlark

- **Суть:** сколько шагов, времени и памяти получает один запуск решателя, сколько запусков идёт сразу, где запускается решатель и что бывает, когда лимит сработал. Решается здесь: песочница — часть `module-kit` (architecture 3.7 spec; R17 spec).
- **Что известно:**
  - в standalone — 25 000 000 шагов и 2 секунды на запуск, по одному запуску на vCPU (architecture 3.7 spec; SPEC 6.6 standalone);
  - там же, по SPEC 6.4–6.6 standalone (раздел 4):
    - исходник решателя — не больше 8 КБ;
    - рекурсия выключена: рекурсивная функция Starlark идёт по стеку Go, а переполнение стека уронило бы весь экземпляр, а не один запрос;
    - встроенные функции платят шаг за каждые 16 байт того, что строят, и не строят списков длиннее 1 000 000 элементов: интерпретатор считает шаги, но не память. Так шаги ограничивают и память, которую держат встроенные функции, — 381 МиБ на потолке шагов; среде Go отдана девять десятых памяти экземпляра (`GOMEMLIMIT`);
    - запуск, который 3 секунды не находит свободного места, не начинается, а задача считается непроверенной;
    - у памяти осталась известная дыра, и она шире одного оператора. Оператор `*` и повторное `xs = xs + xs` строят огромный список мимо счётчиков; литерал `{}` в цикле держит по 512 байт на каждые девять шагов, и экземпляр с 1 ГиБ памяти был убит именно так; методы самих значений — `x.extend(range(3000000))` — берут три миллиона элементов за девятнадцать шагов (SPEC 6.6 standalone; R53 standalone). Внутри одного вызова не видны и часы. Закрыло бы всё это только исполнение решателя в отдельном процессе с лимитом памяти, а v1 standalone на это не пошёл;
    - цену в байтах, один запуск на vCPU и 1 ГиБ памяти на каждый vCPU решил R125 standalone. Отвергнуты там четыре места с часами как страховкой — вердикт зависел бы от нагрузки — и отдельный процесс сразу. Открытым остался отказ в обслуживании: один аккаунт, который сдаёт решатели на все 2 секунды, держит единственное место экземпляра и ничего за это не платит;
    - лимит шагов калибровали по 450 эталонным решателям: худший потребовал 2 081 362 шага, и потолок подняли с 10 000 000 до 25 000 000 (R53 standalone); худший из 603 — 2 413 617 шагов;
  - интерпретатор Starlark в standalone — `go.starlark.net`, закреплённый псевдоверсией коммита (раздел 4);
  - решатель standalone перебором проверял задачу, которую написала модель (SPEC 6.1 standalone). Здесь решатель считает ключ шаблонной задачи и служит виду проверки «условие» (repositories 7 spec; concepts 9.2, 9.3 spec). Порт 603 задач standalone с решателями — в `modules` (repositories 7 spec; T03 modules);
  - где запускается решатель — при проверке модуля, при сборке пакета или в `api` на каждой попытке, — `spec` не говорит. Ключ шаблона зависит от зерна попытки (concepts 9.3 spec), так что посчитать его заранее можно, только если перебрать все зёрна;
  - лимит шагов от машины не зависит, а лимит времени — зависит: тот же решатель на другой машине может не уложиться в 2 секунды. А вердикт повторяется по сохранённому наблюдению (architecture 1 spec; R03 spec);
  - библиотека — вычисления без ввода-вывода (repositories 1 spec). Кто держит часы и считает одновременные запуски — `module-kit` или `api`, — нужно решить;
  - на этапах 1–3 модули только официальные, по рекомендации О-9; с этапа 4 модули публикуют и сторонние авторы, а `module-registry` построен на `module-kit` (R17 spec; architecture 3.13 spec). Лимиты должны выдерживать и чужой решатель. Решатель, которого не остановил лимит памяти, в standalone мог уронить весь экземпляр, а не один запрос (SPEC 6.6 standalone). Здесь это был бы процесс, где решатель исполняется: `api`, если решатель идёт на каждой попытке, или CI `modules` и, с этапа 4, `module-registry`, если при проверке модуля.
- **Варианты:** пока пусто.
- **Задача:** T03.

### 2.4. Правила semver для `diff`

- **Суть:** по каким правилам `module-kit` сравнивает новую версию модуля с прошлой: что поднимает мажорную, минорную и патч-версию, что сравнение находит само, а что объявляет автор. Карточка называет сравнение в том, что `module-kit` делает, но не среди пунктов «Решить при детализации» (architecture 3.7 spec); его называет задача, которая завела этот репозиторий (T05 `spec`). Команды `diff` в карточке тоже нет: CLI там — «проверить, собрать, построить индекс». Отдельная это команда или часть проверки, решает задача.
- **Что известно:**
  - версия модуля — semver. Мажорную поднимает несовместимое изменение: удаление навыка или изменение его смысла, изменение ключа опубликованной задачи. Такое изменение несёт таблицу соответствия старых ID навыков новым, а опубликованный ID навыка не переиспользуется для другого навыка (concepts 3 spec). Что поднимает минорную и патч-версию, `spec` не говорит;
  - удаление навыка и смену ключа видно, сравнив две версии, а «изменение смысла навыка» — нет: как оно попадает в сравнение, решает задача;
  - по таблице соответствия переносят прогресс `api` и `learner-model` после мажорной версии модуля (architecture 3.5, 3.8 spec). Где лежит таблица и в каком она виде, `spec` пока не говорит;
  - старые версии пакетов для повтора вердикта берутся из релизов `modules` (architecture 3.10 spec). Откуда CI `modules` берёт прошлую версию модуля для сравнения, решает T01 modules;
  - что сравнивается — YAML автора или канонический вид пакета — зависит от Т-15 (раздел 2.1);
  - версия модуля — не версия формата. Версию формата называет паспорт (concepts 3 spec), её поднимает изменение поведения вида проверки, по рекомендации Т-18, а правила версий `spec` — Т-11. Здесь речь только о версии модуля.
- **Варианты:** пока пусто.
- **Задача:** T04.

### 2.5. Команда CLI для открытых видов

- **Суть:** нужна ли CLI `module-kit` своя команда, которая отдаёт открытые виды задач сборке страницы образцов `canvas`, или этой сборке хватает команд карточки — проверить, собрать, построить индекс (architecture 3.7 spec). Вопрос R25 spec оставил детализации `module-kit`. Решается здесь: CLI — часть `module-kit` (R17 spec). Новая команда меняет карточку, и её правку итог детализации отправляет строкой таблицы 3 «в реестр spec» (раздел 2).
- **Что известно:**
  - страница образцов — отдельная сборка `canvas` для образцов eval: образец делается той же канвой, что у учеников, и сохраняется файлом на устройстве (architecture 3.2 spec). К первому замеру она без хостинга: сборка скачивается из релиза `canvas` и открывается на устройстве, а задачи и медиафайлы кладутся в неё при сборке, из релиза `modules` через CLI `module-kit`. Во время работы страница по сети не ходит (R25 spec);
  - пакет `canvas` предметов не знает: данные получает только сборка страницы. CI `canvas` берёт CLI `module-kit` и релиз `modules` по точному тегу — это связь «инструмент CI» (R25 spec; repositories 1, 3 spec);
  - CLI только читает и пишет файлы модуля и в сеть не ходит (architecture 3.7 spec): релиз `modules` скачивает сам CI `canvas`;
  - открытый вид — без ключа, решения и ловушек, по списку открытых полей, и строит его `module-kit` (раздел 1; architecture 1 spec; R05 spec). Карточка называет его среди того, что делает библиотека, а у CLI в ней только три команды;
  - у шаблонной задачи числа подставляются от зерна (concepts 9.3 spec), так что и её открытый вид зависит от зерна (раздел 2.2);
  - что лежит в пакете и в какой его части — по рекомендации Т-15 медиафайлы лежат в части для устройства, — а значит, и что такая команда читает, решают T01 и эпик E2 в RUN.md `spec` (раздел 2.1);
  - отдельная команда CLI решается и для сравнения версий — `diff` (раздел 2.4).
- **Варианты:** пока пусто.
- **Задача:** T05.

## 3. Устройство

Сюда переезжает то, что решено по разделу 2: из каких частей компонент состоит, его типы и данные, как он связан с другими репозиториями. Пока здесь только то, с чего начался код (R04).

### 3.1. Код, проверки и релизы

- **Модуль** — `github.com/altnovatrail/module-kit`. В `go.mod` — `go 1.27.0`, нижняя граница для тех, кто его импортирует, а сам Go 1.27.3 закреплён в стадии `toolchain` devcontainer, где `GOTOOLCHAIN=local` (R04).
- **CLI** — `cmd/module-kit`. Пока он отвечает только на `--version`; проверка модуля, сборка пакета и индекса придут с форматом v0 (раздел 1).
- **Проверки** — рецепты `justfile`. Общий `go.yml` `.github` запускает их по одному в job, в стадии `toolchain` (R16 .github):

  | Рецепт | Что делает | Проверка в ruleset |
  |---|---|---|
  | `ci-lint` | gofmt по файлам `.go` модуля, без `testdata/` и `spec/`, и `go vet` | «Go / lint» |
  | `ci-test` | `go test -race -count=1 ./...` | «Go / test» |
  | `ci-vuln` | govulncheck v1.9.0 — известные уязвимости в коде, до которого модуль доходит | «Go / vuln» |
  | `ci-licenses` | go-licenses v2.0.1 — лицензии пакетов, из которых собирается модуль, по списку `ALLOWED_LICENSES` из `go.yml` | «Go / licenses» |

  Инструменты идут через `go run` по точной версии модуля. Их, как и Go в Dockerfile, поднимают руками (R04). Своих исключений лицензий нет.
- **CodeQL** читает и Go: общий `codeql.yml` `.github` v0.1.3 со входом `languages: go` собирает код с Go из стадии `toolchain`, а проверка — «CodeQL / analyze (go)» (R06).
- **Релиз** — `.github/workflows/release.yml` зовёт общий `release.yml` `.github` v0.1.4 по SHA, после зелёного CI на `main` или по кнопке (R07; R18 .github):
  - номер — следующий патч линии `RELEASE_LINE` `0.1`, первый — v0.1.0; считают его рецепты релиза `spec`. Релиз выходит, только если с прошлого изменилось что-то из `paths`: всё, кроме `docs/`, файлов `*.md`, `.github/` и подмодуля `spec`. Правка документов релиза не даёт, а правка кода, `go.mod`, `justfile` или Go в Dockerfile — даёт;
  - файлы собирает рецепт `release-files <версия>` — в стадии `toolchain` выпускаемого коммита, из его дерева без git и подмодулей, в `dist/`, которую он создаёт с нуля:

    | Файл | Что это |
    |---|---|
    | `module-kit_<версия>_<os>_<arch>`, у windows — с `.exe` | CLI для linux, darwin и windows, на amd64 и arm64: `CGO_ENABLED=0`, `-trimpath`, `-buildvcs=false`, `-ldflags "-s -w -X main.version=<версия>"` |
    | `LICENSE` | Apache-2.0 самого `module-kit` |
    | `LICENSE-go` | лицензия Go, BSD-3-Clause: бинарник несёт его стандартную библиотеку |
    | `SHA256SUMS` | суммы sha256 всех файлов выше |

    Сборка воспроизводима: из дерева `git archive` и из рабочей копии она даёт те же суммы;
  - каждый файл общий workflow подписывает cosign без ключей, кладёт рядом его `.sigstore.json`, сверяет подписи до публикации, а после — `gh release verify` и `verify-asset`. Релизы неизменяемые, а теги `v*` держит ruleset 25559103 — условие Т-27;
  - что проверяет у подписи потребитель — Т-27, T37 `spec`.
- **Чего ещё нет:** gopls в образе — Д5 в таблице 4 [RUN.md](../RUN.md).

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
