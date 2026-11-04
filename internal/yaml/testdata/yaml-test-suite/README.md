# The YAML test suite

The cases of the [YAML test suite](https://github.com/yaml/yaml-test-suite) that `TestYAMLTestSuite` and `FuzzRead` read: its data release `data-2022-01-17`, commit `5f49729577242103ae23838ac2ad4d9145aec126`, the one go-yaml v4 tests itself against (`YTS-TAG` in its `GNUmakefile`). Each case is a folder named by its ID, and a case with variants holds one folder for each.

Of each case only these files are kept:

- `in.yaml` — the YAML it reads;
- `in.json` — the JSON value of a case YAML reads, when JSON can hold it;
- `error` — present when YAML 1.2.2 refuses `in.yaml`;
- `===` — the name of the case.

The suite is under the MIT License, in [License](License), taken from the suite's source at its tag `v2022-01-17`, commit `6d48918a2320e767f6e2e57304f5ab42c19d71db`. The files are byte for byte the release's: they are test data, so no check of this repository reads or rewrites them, and they are not in a release of module-kit.
