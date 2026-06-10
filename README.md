This is a small tool to reduce the size of the NOTICE file created by the `licensed` tool with the command `licensed notice`.

## Usage

`licensed-notice-deduplicate` reads a NOTICE file produced by `licensed notice` and
replaces it with a deduplicated version in place.

All entries that share exactly the same license text are merged into a single
block that lists every covered package, regardless of whether they belong to
the same repository or not.

Example:

`licensed-notice-deduplicate .licenses/NOTICE.project`

## License

This tool is released under GPL-3.0 or later.

```
SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
SPDX-License-Identifier: GPL-3.0-or-later
```
