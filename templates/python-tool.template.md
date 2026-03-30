# python-tool.template.md
# Template-For: python-tool
# Version: 0.1.0
# Spec-Schema: 0.3.19
# LANGUAGE: python

---

## OVERVIEW

A Python command-line tool. Single executable entry point, src/ layout,
fully typed, packaged for RPM, DEB, PyPI, and OCI.

---

## META

```
Name:           {TOOL_NAME}
Version:        {VERSION}
Deployment:     python-tool
Spec-Schema:    0.3.19
Author:         {AUTHOR_NAME} <{AUTHOR_EMAIL}>
License:        {LICENSE}
Python:         >=3.11
Verification:   none
Safety-Level:   QM
```

---

## PROJECT STRUCTURE

```
{TOOL_NAME}/
├── pyproject.toml
├── LICENSE
├── README.md
├── src/
│   └── {PACKAGE_NAME}/
│       ├── __init__.py
│       ├── __main__.py
│       ├── cli.py
│       └── {CORE_MODULE}.py
├── tests/
│   ├── __init__.py
│   ├── test_{CORE_MODULE}.py
│   └── test_cli.py
├── packaging/
│   ├── {TOOL_NAME}.spec
│   └── debian/
│       ├── control
│       ├── changelog
│       ├── rules
│       └── copyright
├── Containerfile
└── Makefile
```

INVARIANTS:
- src/ layout is mandatory — no flat layout
- One package under src/ — name derived from TOOL_NAME
  (hyphens replaced with underscores)
- __main__.py enables: python -m {PACKAGE_NAME}
- cli.py owns all argparse logic; no argument parsing in other modules
- Business logic lives in separate modules, never in cli.py
- tests/ mirrors src/ structure

---

## TOOLCHAIN

```
Runtime:   python >= 3.11
Manager:   uv
Linting:   flake8
Format:    black
Types:     mypy (strict)
Testing:   pytest + hypothesis
```

EXECUTION:
```
# Setup
uv sync

# Run
uv run {TOOL_NAME}

# Lint + format
uv run flake8 src/ tests/
uv run black src/ tests/

# Type check
uv run mypy src/

# Test
uv run pytest tests/

# Build wheel
uv build
```

---

## pyproject.toml

```toml
[project]
name = "{TOOL_NAME}"
version = "{VERSION}"
description = "{DESCRIPTION}"
readme = "README.md"
license = { text = "{LICENSE}" }
authors = [{ name = "{AUTHOR_NAME}", email = "{AUTHOR_EMAIL}" }]
requires-python = ">=3.11"
dependencies = [
    # {DEPENDENCIES}
]

[project.scripts]
{TOOL_NAME} = "{PACKAGE_NAME}.cli:main"

[project.urls]
Homepage = "{HOMEPAGE}"
Repository = "{REPOSITORY}"

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[tool.hatch.build.targets.wheel]
packages = ["src/{PACKAGE_NAME}"]

[tool.mypy]
strict = true
python_version = "3.11"

[tool.pytest.ini_options]
testpaths = ["tests"]

[dependency-groups]
dev = [
    "flake8>=7.0",
    "black>=24.0",
    "mypy>=1.0",
    "pytest>=8.0",
    "hypothesis>=6.0",
]
```

---

## SOURCE FILES

### src/{PACKAGE_NAME}/__init__.py

```python
# SPDX-License-Identifier: {LICENSE}
# SPDX-FileCopyrightText: {YEAR} {AUTHOR_NAME} <{AUTHOR_EMAIL}>

"""
{TOOL_NAME} -- {DESCRIPTION}
"""

__version__: str = "{VERSION}"
```

### src/{PACKAGE_NAME}/__main__.py

```python
# SPDX-License-Identifier: {LICENSE}
# SPDX-FileCopyrightText: {YEAR} {AUTHOR_NAME} <{AUTHOR_EMAIL}>

"""Enable: python -m {PACKAGE_NAME}"""

from {PACKAGE_NAME}.cli import main

if __name__ == "__main__":
    main()
```

### src/{PACKAGE_NAME}/cli.py

```python
# SPDX-License-Identifier: {LICENSE}
# SPDX-FileCopyrightText: {YEAR} {AUTHOR_NAME} <{AUTHOR_EMAIL}>

"""CLI entry point -- argument parsing only. No business logic here."""

import argparse
import logging
import sys

from {PACKAGE_NAME} import __version__
from {PACKAGE_NAME}.{CORE_MODULE} import run


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="{TOOL_NAME}",
        description="{DESCRIPTION}",
    )
    parser.add_argument(
        "--version",
        action="version",
        version=f"%(prog)s {__version__}",
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="Enable verbose logging",
    )
    # {ADDITIONAL_ARGUMENTS}
    return parser


def main() -> None:
    parser = build_parser()
    args = parser.parse_args()

    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(levelname)s %(name)s: %(message)s",
        stream=sys.stderr,
    )

    try:
        run(args)
    except KeyboardInterrupt:
        sys.exit(130)
    except Exception as exc:
        logging.getLogger(__name__).error("%s", exc)
        sys.exit(1)
```

### src/{PACKAGE_NAME}/{CORE_MODULE}.py

```python
# SPDX-License-Identifier: {LICENSE}
# SPDX-FileCopyrightText: {YEAR} {AUTHOR_NAME} <{AUTHOR_EMAIL}>

"""Core business logic -- no CLI concerns here."""

import argparse
import logging

logger = logging.getLogger(__name__)


def run(args: argparse.Namespace) -> None:
    """Entry point called by cli.main()."""
    logger.debug("starting with args=%r", args)
    # {IMPLEMENTATION}
```

---

## TESTING

### tests/test_{CORE_MODULE}.py

```python
# SPDX-License-Identifier: {LICENSE}
# SPDX-FileCopyrightText: {YEAR} {AUTHOR_NAME} <{AUTHOR_EMAIL}>

"""Tests for {CORE_MODULE}. Use hypothesis for property-based tests."""

import argparse

from hypothesis import given, strategies as st

from {PACKAGE_NAME}.{CORE_MODULE} import run


def test_run_smoke() -> None:
    args = argparse.Namespace(verbose=False)
    run(args)  # must not raise


# Example property-based test -- adapt or remove
@given(st.text())
def test_run_with_input(value: str) -> None:
    args = argparse.Namespace(verbose=False, input=value)
    run(args)
```

---

## PACKAGING

### RPM: packaging/{TOOL_NAME}.spec

```spec
Name:           {TOOL_NAME}
Version:        {VERSION}
Release:        1%{?dist}
Summary:        {DESCRIPTION}

License:        {LICENSE}
URL:            {HOMEPAGE}
Source0:        %{name}-%{version}.tar.gz

BuildArch:      noarch
BuildRequires:  python3 >= 3.11
BuildRequires:  python3-pip
BuildRequires:  python3-build
Requires:       python3 >= 3.11
# {RUNTIME_DEPENDENCIES}

%description
{DESCRIPTION}

%prep
%autosetup

%build
python3 -m build --wheel --no-isolation

%install
pip install --no-deps --root %{buildroot} \
    dist/%{name}-%{version}-py3-none-any.whl

%files
%license LICENSE
%{_bindir}/{TOOL_NAME}
%{python3_sitelib}/{PACKAGE_NAME}/
%{python3_sitelib}/{PACKAGE_NAME}-%{version}.dist-info/

%changelog
* {DATE} {AUTHOR_NAME} <{AUTHOR_EMAIL}> - {VERSION}-1
- Initial release
```

### DEB: packaging/debian/control

```
Source: {TOOL_NAME}
Section: utils
Priority: optional
Maintainer: {AUTHOR_NAME} <{AUTHOR_EMAIL}>
Build-Depends: debhelper-compat (= 13),
               dh-python,
               python3-all,
               python3-setuptools
Standards-Version: 4.6.1
Homepage: {HOMEPAGE}

Package: {TOOL_NAME}
Architecture: all
Depends: python3 (>= 3.11), ${python3:Depends}, ${misc:Depends}
Description: {DESCRIPTION}
 {LONG_DESCRIPTION}
```

### OCI: Containerfile

```dockerfile
FROM registry.opensuse.org/opensuse/leap:15.6

RUN zypper --non-interactive install --no-recommends \
    python311 \
    && zypper clean --all

WORKDIR /app
COPY dist/{TOOL_NAME}-{VERSION}-py3-none-any.whl .

RUN pip3 install --no-cache-dir \
    {TOOL_NAME}-{VERSION}-py3-none-any.whl

ENTRYPOINT ["{TOOL_NAME}"]
```

---

## MAKEFILE

```makefile
.PHONY: all lint format typecheck test build clean

all: lint typecheck test build

lint:
	uv run flake8 src/ tests/

format:
	uv run black src/ tests/

typecheck:
	uv run mypy src/

test:
	uv run pytest tests/ -v

build:
	uv build

clean:
	rm -rf dist/ .mypy_cache/ .pytest_cache/ __pycache__
```

---

## VARIABLES

| Variable | Required | Description |
|---|---|---|
| TOOL_NAME | yes | Hyphenated tool name, e.g. my-tool |
| PACKAGE_NAME | yes | Python package name (underscores), e.g. my_tool |
| CORE_MODULE | yes | Main logic module name, e.g. core |
| VERSION | yes | Semantic version, e.g. 0.1.0 |
| DESCRIPTION | yes | One-line description |
| LONG_DESCRIPTION | yes | Multi-line description for DEB |
| AUTHOR_NAME | yes | Full name |
| AUTHOR_EMAIL | yes | Email address |
| LICENSE | yes | SPDX identifier, e.g. GPL-2.0-only |
| YEAR | yes | Copyright year |
| HOMEPAGE | yes | Project URL |
| REPOSITORY | yes | VCS URL |
| DEPENDENCIES | no | Runtime deps as TOML list entries |
| RUNTIME_DEPENDENCIES | no | RPM Requires: lines |
| ADDITIONAL_ARGUMENTS | no | Extra argparse add_argument() calls |
| IMPLEMENTATION | no | Initial body of run() |
| DATE | yes | RPM changelog date |

---

## INVARIANTS

- SPDX header in every .py file — mandatory, no exceptions
- LICENSE file present at repo root — mandatory
- pyproject.toml carries license field — mandatory
- cli.py contains only argparse setup and main() — no business logic
- All public functions and methods carry type annotations — mypy strict
- flake8 and black pass with zero warnings before any commit
- mypy --strict passes with zero errors before any commit
- Vendoring is optional — document in README if used
- OCI base image: openSUSE Leap current release — no Alpine, no Debian
- Wheel built with uv build or python3 -m build — not setup.py
- pip install with --no-deps in RPM %install — distro manages deps
- curl | sh installation: forbidden
