envi
====

[![GoDoc](https://godoc.org/github.com/Kochava/envi?status.svg)](https://godoc.org/github.com/Kochava/envi)
[![CircleCI](https://circleci.com/gh/Kochava/envi.svg?style=svg)](https://circleci.com/gh/Kochava/envi)
[![codecov](https://codecov.io/gh/Kochava/envi/branch/master/graph/badge.svg)](https://codecov.io/gh/Kochava/envi)


envi is a library for loading data from structured environment variables and
maps of strings (e.g., envi.Values) into structs, slices, and general scalars
(strings, ints, bools, etc.).

Currently, you can unmarshal into the following types out of the box:

- Integers (minus uintptr)
- Floats
- Bools
- time.Duration
- Strings and byte slices (treated equivalently)
- url.URL
- Anything that implements encoding.TextUnmarshaler or envi.Unmarshaler
- Slices of the above
- Structs with exported fields of the above (including nested structs)


License
-------

envi is licensed under a BSD-3 clause license, the same as Go and can also be
read in the LICENSE.txt file:

> Copyright (c) 2015 Kochava.
> All rights reserved.
>
> Redistribution and use in source and binary forms, with or without
> modification, are permitted provided that the following conditions are
> met:
>
> 1. Redistributions of source code must retain the above copyright
>    notice, this list of conditions and the following disclaimer.
>
> 2. Redistributions in binary form must reproduce the above copyright
>    notice, this list of conditions and the following disclaimer in the
>    documentation and/or other materials provided with the distribution.
>
> 3. Neither the name of the copyright holder nor the names of its
>    contributors may be used to endorse or promote products derived from
>    this software without specific prior written permission.
>
> THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS
> IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED
> TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
> A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
> HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
> SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
> TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
> PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
> LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
> NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
> SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

Some code in envi, specifically code under indirect.go, is borrowed and modified
from the Go standard library and is under Go's license, reproduced below:

> Copyright (c) 2009 The Go Authors. All rights reserved.
>
> Redistribution and use in source and binary forms, with or without
> modification, are permitted provided that the following conditions are
> met:
>
>    * Redistributions of source code must retain the above copyright
> notice, this list of conditions and the following disclaimer.
>    * Redistributions in binary form must reproduce the above
> copyright notice, this list of conditions and the following disclaimer
> in the documentation and/or other materials provided with the
> distribution.
>    * Neither the name of Google Inc. nor the names of its
> contributors may be used to endorse or promote products derived from
> this software without specific prior written permission.
>
> THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
> "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
> LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
> A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
> OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
> SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
> LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
> DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
> THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
> (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
> OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.


## Default Branch

As of October 1, 2020, github.com uses the branch name ‘main’ when creating the initial default branch for all new repositories.  In order to minimize any customizations in our github usage and to support consistent naming conventions, we have made the decision to rename the ‘master’ branch to be called ‘main’ in all Kochava’s github repos.

For local copies of the repo, the following steps will update to the new default branch:

```
git branch -m master main
git fetch origin
git branch -u origin/main main
git remote set-head origin -a
```
