<!--
Copyright 2025-2026 Stanislav Senotrusov

This work is dual-licensed under the Apache License, Version 2.0
and the MIT License. Refer to the LICENSE file in the top-level directory
for the full license terms.

SPDX-License-Identifier: Apache-2.0 OR MIT
-->

## Installation

Build and install all commands from `./cmd` into `$GOPATH/bin`:

```sh
go install ./cmd/...
```

You probably need to install [term-clipboard](https://github.com/senotrusov/term-clipboard) as well (depending on your use case, see [How to use it](#how-to-use-it) below).

## How to use it

Copy [example-config/darntext](example-config/darntext) to your [user config directory](https://pkg.go.dev/os#UserConfigDir) and customize it as needed.

The `# dir: ~/example` comment in [example.sh](example-config/darntext/example.sh), located immediately above the script body, specifies the project directory to which the configuration applies.

You can define multiple `# dir:` lines to apply the same configuration to multiple projects.

After updating the directory paths, change to one of the configured project directories and run:

```sh
darntext
````

If the configuration is loaded successfully, `darntext` lists the available commands, for example:

* `apply`
* `context`
* `task`
* `tidy`
* `re`

Run a command by passing its name to `darntext`:

```sh
darntext context
```

## License

This work is dual-licensed under the Apache License, Version 2.0
and the MIT License. Refer to the [LICENSE](LICENSE) file in the top-level
directory for the full license terms.

## Get involved

See the [CONTRIBUTING](CONTRIBUTING.md) file for guidelines
on how to contribute, and the [CONTRIBUTORS](CONTRIBUTORS.md)
file for a list of contributors.
