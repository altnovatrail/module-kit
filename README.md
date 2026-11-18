# module-kit

[![CI](https://github.com/altnovatrail/module-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/altnovatrail/module-kit/actions/workflows/ci.yml)
[![Release](https://github.com/altnovatrail/module-kit/actions/workflows/release.yml/badge.svg)](https://github.com/altnovatrail/module-kit/actions/workflows/release.yml)
[![CodeQL](https://github.com/altnovatrail/module-kit/actions/workflows/codeql.yml/badge.svg)](https://github.com/altnovatrail/module-kit/actions/workflows/codeql.yml)
[![Latest release](https://img.shields.io/github/v/release/altnovatrail/module-kit)](https://github.com/altnovatrail/module-kit/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/altnovatrail/module-kit)](go.mod)
[![Coverage](https://codecov.io/gh/altnovatrail/module-kit/branch/main/graph/badge.svg)](https://codecov.io/gh/altnovatrail/module-kit)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/16385/badge)](https://www.bestpractices.dev/projects/16385)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/altnovatrail/module-kit/badge)](https://scorecard.dev/viewer/?uri=github.com/altnovatrail/module-kit)
[![License](https://img.shields.io/github/license/altnovatrail/module-kit)](LICENSE)

[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=altnovatrail_module-kit&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=altnovatrail_module-kit&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=altnovatrail_module-kit&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=altnovatrail_module-kit&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=altnovatrail_module-kit&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit)

The NovaTrail module format in code: the one Go library that holds the format's semantics, and the authors' CLI on top of it. It checks an author's module, builds its package and the module index, compares module versions, and fills seeded task templates and runs their solvers in a Starlark sandbox; for the platform's service it builds the open view of a task, gives the verdict by the check kinds, runs the leak check and trusts a package by the chain of digests from a pinned index. It never calls an AI model, makes no network calls and holds no learners' data.

A component of [NovaTrail](https://github.com/altnovatrail/spec), a universal open platform for adaptive learning, interactive lessons and AI tutoring. The format itself — its schemas, rules and test vectors — is spec's, and the library implements it and passes every vector; it depends on nothing but spec.

- `api` runs it in the service's process: it hands the library the observation that `tutor` read from a learner's work and takes back the verdict by key or predicate and the open view that the learner's app receives;
- tutor's eval runner imports the library too;
- the CI of `modules` checks and builds the modules with its CLI, and the CI of `canvas` builds the samples page with it;
- module-registry and module-pipeline will build on it later, while `sim` reads module packages through spec's types without it.

Its card is section 3.7 of spec's [architecture](https://github.com/altnovatrail/spec/blob/main/docs/architecture.md); the detail is under [docs/](docs/README.md): what the component does and how it is built, its decisions, its releases and how a change comes in.
