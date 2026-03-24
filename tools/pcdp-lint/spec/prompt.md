
I am providing two files:

1. cli-tool.template.md — a deployment template that defines
   the conventions, constraints, and defaults for this type of component
   under the Post-Coding Development Paradigm.

2. pcdp-lint.md — a specification for a component, written in the
   Post-Coding Development Paradigm format.

Your task:
Implement the component in full, exactly as specified. Do not add
features not described in the specification. Do not omit any
specified behaviour.

Target language is Go (derived from the cli-tool template default).

## Resuming a partial run

Before writing any file, read the output directory. If any of the
files listed below already exist and are non-empty, do not overwrite
them — treat them as complete and move to the next missing file.
Report which files you found and which you are producing.

## Required deliverables and delivery order

Produce files in this exact order. Complete each file fully before
starting the next.

Phase 1 — Core implementation:
  main.go
  go.mod

Phase 2 — Build and packaging:
  Makefile
  pcdp-lint.spec
  debian/control
  debian/changelog
  debian/rules
  debian/copyright
  LICENSE

Phase 3 — Documentation and report (last):
  README.md
  TRANSLATION_REPORT.md

Do not produce TRANSLATION_REPORT.md until all other files are
written and verified on disk. If you are interrupted, restart from
the first missing file in the order above.

## LICENSE

Never write the full GPL-2.0-only license text to disk. Write only:
name, one-line description, and a reference to the authoritative
source. The full text is managed separately.

## Delivery

You have access to a filesystem MCP server. Write all files directly
to disk. Report each file path as you complete it.

Do not attempt to compile, execute, or install anything unless
explicitly asked.

## Translation report

TRANSLATION_REPORT.md must cover:
- Target language and why (template default, no preset override)
- Delivery mode used
- How STEPS ordering was followed for each BEHAVIOR block
- Any specification ambiguities encountered
- Any rules you could not implement exactly as written, and why
- Your confidence per EXAMPLE as a table with these exact columns:

  | EXAMPLE | Confidence | Verification method | Unverified claims |

  A claim is verified only if it references a specific named test
  function that passes without a live external service.
  Unverified claims must be listed explicitly — never silently omitted.

Do not ask clarifying questions. If the specification is ambiguous,
make the most conservative interpretation, implement it, and note
the ambiguity in the translation report.
