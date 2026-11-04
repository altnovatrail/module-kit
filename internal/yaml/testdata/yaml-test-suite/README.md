# The YAML test suite

The cases of the [YAML test suite](https://github.com/yaml/yaml-test-suite) that `TestYAMLTestSuite` and `FuzzRead` read: its data release `data-2022-01-17`, commit `5f49729577242103ae23838ac2ad4d9145aec126`, the one go-yaml v4 tests itself against (`YTS-TAG` in its `GNUmakefile`).

The release holds each case as a folder named by its ID, and a case with variants as a folder for each, such as `DK95/01`. Here they are grouped into two files, so that 402 cases are not 1180 files:

- `valid.json` — the 308 cases YAML 1.2.2 reads;
- `error.json` — the 94 cases it refuses, those whose folder holds the file `error`.

Each file is a JSON array of the cases, in the order of their IDs, and each case an object:

- `id` — the path of its folder in the release;
- `name` — its file `===`, without the line break at its end;
- `yaml` — its file `in.yaml`, as it is;
- `json` — its file `in.json`, as it is, when the case has one: the JSON value of what YAML reads, when JSON can hold it.

A JSON string holds the text of a file exactly, tabs, line breaks and every other character escaped as JSON needs, and every `in.yaml` of the release is UTF-8. The release's folder `tags/`, which lists the cases by topic through symbolic links, is left out.

The suite is under the MIT License, in [License](License), taken from the suite's source at its tag `v2022-01-17`, commit `6d48918a2320e767f6e2e57304f5ab42c19d71db`. These files are test data: no check of this repository rewrites them, and they are not in a release of module-kit.
