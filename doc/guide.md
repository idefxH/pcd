# PCD User Guide — Writing Specifications

This guide is for domain experts and engineers who want to use Post-Coding
Development (PCD) to specify and generate software. You do not need to know
any programming language or formal notation to write a valid PCD specification.

**The key rule:** you write what the component should do. The AI translator
writes how. If the generated code is wrong, you fix the specification and
regenerate — you never edit the code directly.

---

## Contents

1. [Workflow overview](#1-workflow-overview)
2. [Writing your first specification](#2-writing-your-first-specification)
3. [Specification structure reference](#3-specification-structure-reference)
4. [Milestones — phased translation for large components](#4-milestones)
5. [Hints files](#5-hints-files)
6. [Language neutrality](#6-language-neutrality)
7. [Translating to code](#7-translating-to-code)
8. [Reverse-engineering an existing codebase](#8-reverse-engineering)
9. [Prompts reference](#9-prompts-reference)
10. [Validating your specification](#10-validating-your-specification)

---

## 1. Workflow Overview

```
You write a spec  →  pcd-lint validates it  →  AI translator generates code
       ↑                                              |
       └──────── fix the spec if output is wrong ─────┘
```

For large components, translation is split into milestones:

```
Spec (all BEHAVIORs)
  │
  ├── MILESTONE: 0.0.0  Scaffold: true   → all files, all stubs, compile gate
  ├── MILESTONE: 0.1.0                   → implement group A
  ├── MILESTONE: 0.2.0                   → implement group B
  └── MILESTONE: 0.3.0                   → implement group C
```

A pipeline agent can advance milestones automatically; you only intervene
on failures.

---

## 2. Writing Your First Specification

### Option A — AI-assisted interview (recommended)

You do not need to learn the spec format first. Use the interview prompt with
any capable LLM:

```bash
# with a local model:
ollama run llama3.2 "$(cat prompts/interview-prompt.md)"

# or paste prompts/interview-prompt.md as the system prompt in any chat interface
```

The model will ask you questions one at a time and produce a complete spec at
the end. It handles both new components (full interview) and existing material
(gap-fill from notes, emails, design docs).

### Option B — Reverse-engineering existing code

If you have existing source code you want to analyse, refactor, or port:

```bash
# paste prompts/reverse-prompt.md as the system prompt, then share your source code
```

The model reads the code, extracts the spec structure, confirms the deployment
type and target language with you, and asks what changes you want to make.

### Option C — Write directly

Every spec follows this skeleton. Fill in the sections described in Section 3.

```markdown
# My Component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.3.21
Author:       Your Name <you@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES
...

## BEHAVIOR: my-operation
Constraint: required
INPUTS: ...
PRECONDITIONS: ...
STEPS:
  1. [action]; on failure → [error].
  2. [next action].
POSTCONDITIONS: ...
ERRORS: ...

## PRECONDITIONS
...

## POSTCONDITIONS
...

## INVARIANTS
- [observable]      ...
- [implementation]  ...

## EXAMPLES

EXAMPLE: success_case
GIVEN: ...
WHEN:  ...
THEN:  ...

EXAMPLE: error_case
GIVEN: ...
WHEN:  ...
THEN:  result = Err(...)
```

Validate with `pcd-lint myspec.md` before translating.

---

## 3. Specification Structure Reference

### Required sections

All seven must be present or pcd-lint will report errors.

#### `## META`

```markdown
## META
Deployment:   <template>       # cli-tool | mcp-server | cloud-native | ...
Version:      0.1.0            # your spec version (MAJOR.MINOR.PATCH)
Spec-Schema:  0.3.21           # PCD schema version — use current
Author:       Name <email>     # repeatable; multiple Author: lines allowed
License:      Apache-2.0       # SPDX identifier
Verification: none             # none | lean4 | fstar | dafny | custom
Safety-Level: QM               # QM | ASIL-A | ASIL-B | ...
```

**Deployment type** determines the target language, packaging, and build
conventions automatically. You never declare a language in the spec.

| Deployment type | What it is |
|---|---|
| `cli-tool` | Command-line tool, single binary |
| `mcp-server` | MCP server (stdio + HTTP transports) |
| `cloud-native` | Kubernetes operator or controller |
| `backend-service` | Long-running Linux service (systemd) |
| `gui-tool` | Desktop/mobile GUI application |
| `python-tool` | Python CLI or automation script |
| `library-c-abi` | General-purpose C-ABI library |
| `verified-library` | Safety/security-critical C library with formal verification |
| `project-manifest` | Multi-component project definition |

#### `## TYPES`

Declare all data types the component works with. Use pseudocode notation —
no programming language syntax.

```markdown
## TYPES

```
Account := {
  id:      string where non-empty,
  balance: int    where balance >= 0
}

TransferResult := Ok | Err(ErrorCode)

ErrorCode := InsufficientFunds | InvalidAccount | SameAccount
```
```

#### `## BEHAVIOR: {name}`

One block per operation. Every block must have INPUTS, PRECONDITIONS, STEPS,
POSTCONDITIONS, and ERRORS.

```markdown
## BEHAVIOR: transfer
Constraint: required

INPUTS:
```
from:   Account
to:     Account
amount: int where amount > 0
```

PRECONDITIONS:
- from.id ≠ to.id
- from.balance >= amount

STEPS:
1. Validate that from.id ≠ to.id; on failure → ERR_SAME_ACCOUNT.
2. Validate that from.balance >= amount; on failure → ERR_INSUFFICIENT_FUNDS.
3. Deduct amount from from.balance.
4. Add amount to to.balance.

POSTCONDITIONS:
- from.balance decreased by amount
- to.balance increased by amount
- sum of all balances is unchanged

ERRORS:
- ERR_SAME_ACCOUNT if from.id = to.id
- ERR_INSUFFICIENT_FUNDS if from.balance < amount
```

**STEPS rules:**
- Numbered, imperative sentences
- Every step that can fail must say: `on failure → [error action]`
- Use `MECHANISM:` annotation when the *how* matters for correctness,
  not just the *what*

**Constraint values:**
- `required` (default): always implement
- `supported`: implement only if active in the resolved preset
- `forbidden`: never implement (add a `reason:` annotation)

Use `## BEHAVIOR/INTERNAL: {name}` for implementation logic not directly
exposed to users. Same structural rules apply.

#### `## PRECONDITIONS`

Global preconditions that apply before the component can run at all.

#### `## POSTCONDITIONS`

Global postconditions guaranteed after any successful operation.

#### `## INVARIANTS`

Rules that must always hold. Tag each entry:

```markdown
## INVARIANTS

- [observable]      sum of all account balances never changes across transfers
- [implementation]  SSH private key bytes are never written to any file path
```

- `[observable]` — verifiable by external observation or the test suite
- `[implementation]` — verifiable only by code review or static analysis

#### `## EXAMPLES`

At least one complete example. Every BEHAVIOR with error exits in STEPS
must have at least one negative-path example (THEN shows an error outcome).

```markdown
## EXAMPLES

EXAMPLE: successful_transfer
GIVEN:
  account A has balance 100
  account B has balance 50
WHEN:
  transfer(A, B, 30)
THEN:
  A.balance = 70
  B.balance = 80
  result = Ok

EXAMPLE: insufficient_funds
GIVEN:
  account A has balance 20
WHEN:
  transfer(A, B, 50)
THEN:
  result = Err(ERR_INSUFFICIENT_FUNDS)
  A.balance unchanged
  B.balance unchanged
```

Multi-pass examples (for reconcilers, retry loops, etc.):

```markdown
EXAMPLE: graceful_stop
GIVEN:
  VM "test-01" is Running
  desiredState = Stopped
WHEN:  reconcile runs (pass 1)
THEN:
  domain.Shutdown() called
  result = RequeueAfter(10s)

WHEN:  reconcile runs (pass 2); domain now Shutoff
THEN:
  status.phase = Stopped
  result = RequeueAfter(60s)
```

---

### Optional sections

#### `## INTERFACES`

Declare external system boundaries and their test doubles. This prevents the
translator from making ad-hoc abstraction decisions and keeps tests
infrastructure-free.

```markdown
## INTERFACES

Store {
  required-methods:
    Load(id string) → (Record, error)
    Save(r Record) → error
  implementations-required:
    production:  PostgresStore
    test-double: FakeStore {
      configurable fields: records map[string]Record, loadErr error
    }
}
```

#### `## DEPENDENCIES`

Declare external library requirements. The translator must not fabricate
version strings or commit hashes.

```markdown
## DEPENDENCIES

github.com/some/library:
  minimum-version: v1.2.3
  rationale: required for X feature
  do-not-fabricate: true
  hints-file: cli-tool.go.some-library.hints.md
```

#### `## TOOLCHAIN-CONSTRAINTS`

Spec-specific overrides for OCI builds, generated files, or other toolchain
constraints that the deployment template does not cover.

#### `## DELIVERABLES`

For multi-component projects: declare logical COMPONENT entries. The
translator maps these to concrete filenames via the deployment template.

#### `## DELTA`

A single-pass work order for the next translation. Lists changes not yet
reflected in the BEHAVIOR sections. Ephemeral — removed after a successful
translation pass.

```markdown
## DELTA

- Add --json output flag to the list subcommand
- Fix error handling in the transport layer (currently silently drops errors)
```

#### `## MILESTONE`

For large components: defines phased translation. See Section 4.

---

## 4. Milestones

Use milestones when your component is too large to translate in one pass —
roughly when you have more than 10 BEHAVIORs or expect more than 500 lines
of generated code.

### The scaffold-first pattern

The first milestone should always be a scaffold pass: it creates all files,
all types, all function signatures, and all stub bodies for the entire
component. The only acceptance criterion is a clean compile. All subsequent
milestones fill in real implementations without touching the file structure.

This has been validated empirically: a 35-BEHAVIOR specification produced
~1600 lines of Go scaffold in one translator session, and the scaffold held
without modification through seven subsequent implementation milestones. The
same spec was also translated to Rust in one session with the same result.

### Milestone syntax

```markdown
## MILESTONE: 0.0.0
Status: pending
Scaffold: true
Hints-file: cli-tool.go.milestones.hints.md, mycomponent.implementation.hints.md

Included BEHAVIORs:
  operation-a, operation-b, operation-c, operation-d, operation-e

Acceptance criteria:
  ./mycomponent --version | grep -q "^mycomponent "
  ./mycomponent --help | grep -q "usage:"

## MILESTONE: 0.1.0
Status: pending

Included BEHAVIORs:
  operation-a, operation-b

Deferred BEHAVIORs:
  operation-c, operation-d, operation-e

Acceptance criteria:
  ./mycomponent run | jq '.result | length > 0' | grep -q true
  test -s /tmp/output.json

## MILESTONE: 0.2.0
Status: pending

Included BEHAVIORs:
  operation-c, operation-d

Deferred BEHAVIORs:
  operation-e

Acceptance criteria:
  ./mycomponent full | jq '.items | length > 3' | grep -q true
```

### Status values

| Status | Meaning | Set by |
|---|---|---|
| `pending` | Not yet attempted | Author (initial) |
| `active` | Currently being translated | Agent pipeline |
| `failed` | Gates did not pass | Agent pipeline |
| `released` | All gates passed, frozen | Agent pipeline |

Exactly one milestone may be `active` at a time. The pipeline agent
(`mcp-server-pcd` `set_milestone_status` tool) advances the cursor;
you only intervene on failures.

### Field reference

| Field | Required | Description |
|---|---|---|
| `Status:` | Yes | Pipeline state: pending/active/failed/released |
| `Scaffold:` | No | `true` = scaffold pass (default: false) |
| `Hints-file:` | No | Comma-separated hints files to read before translating |
| `Included BEHAVIORs:` | Yes | BEHAVIORs to implement fully in this milestone |
| `Deferred BEHAVIORs:` | No (omit for scaffold) | BEHAVIORs to leave as stubs |
| `Acceptance criteria:` | Recommended | Shell commands; exit 0 = pass |

### Rules

- At most one `Scaffold: true` milestone per spec (RULE-17)
- The scaffold milestone must appear first in document order (RULE-17)
- Every BEHAVIOR listed in Included or Deferred must exist in the spec (RULE-16)
- All BEHAVIOR names in the spec must appear in some milestone's Included
  or Deferred list (or they are always translated in full with no phasing)

### Acceptance criteria format

Write criteria as shell commands that exit 0 on pass and non-zero on failure.
This makes them automatable:

```
./sitar version | grep -q "^sitar "
./sitar all outdir=/tmp/test && test -s /tmp/test/general.json
jq '.cpu._elements | length > 0' /tmp/test/json/cpu.json | grep -q true
```

For components requiring elevated privileges, M0 criteria must be runnable
without privilege. M1+ criteria may require a privileged environment and
should be phrased so the human verifier knows what to run.

---

## 5. Hints Files

Hints files contain implementation knowledge that belongs in neither the spec
(which must be language-agnostic) nor the template (which covers conventions,
not library internals). They are advisory only and cannot override spec
invariants.

**Three layers:**

| File pattern | Contents |
|---|---|
| `<template>.<lang>.milestones.hints.md` | Scaffold patterns, stub conventions, file layout, compile gate commands. Reusable across all components using this template and language. |
| `<component>.implementation.hints.md` | Component-specific, language-neutral. File groupings, required field names, known failure modes from prior runs. |
| `<template>.<lang>.<library>.hints.md` | Library-specific API shapes, verified version strings, known gotchas. |

Reference hints files from your spec via DEPENDENCIES, or from a MILESTONE
via the `Hints-file:` field.

---

## 6. Language Neutrality

A spec that may be translated into more than one language must contain no
language-specific constructs anywhere in TYPES, BEHAVIOR, INTERFACES,
INVARIANTS, or MILESTONE acceptance criteria.

**Right:**
```
STEPS:
  1. Create the output directory recursively if it does not exist.
  2. Write the result as JSON to {outdir}/result.json.
```

**Wrong:**
```
STEPS:
  1. Call os.MkdirAll(outdir, 0755).
  2. json.Marshal the result and write to outdir/result.json.
```

**Acceptance criteria — right:**
```
test -d /tmp/out && test -s /tmp/out/result.json
jq '.status' /tmp/out/result.json | grep -q '"ok"'
```

**Acceptance criteria — wrong:**
```
go build ./... && ./mytool run
cargo test --release
```

A useful test: can a developer who knows the domain but has not decided on a
target language read and understand this section? If not, language-specific
content has leaked into the spec.

---

## 7. Translating to Code

Once your spec passes `pcd-lint` with zero errors:

```bash
pcd-lint myspec.md   # must show ✓ before proceeding
```

Use `prompts/prompt.md` as the system prompt with any capable LLM, then
provide the spec and the appropriate deployment template:

```
Input files in the same directory:
  cli-tool.template.md    (the deployment template)
  myspec.md               (your specification)
```

The translator will:
1. Derive the target language from the template
2. Follow the template's EXECUTION phases
3. Produce all required deliverables
4. Write a `TRANSLATION_REPORT.md` documenting every decision

**If a milestone is active**, the translator reads the active milestone,
implements only its Included BEHAVIORs, generates stubs for Deferred
BEHAVIORs, and verifies the acceptance criteria.

**If the output is wrong**, fix the spec and regenerate. Do not edit
the generated code.

---

## 8. Reverse-Engineering Existing Code

Use `prompts/reverse-prompt.md` to produce a PCD spec from existing source:

1. Paste `reverse-prompt.md` as the system prompt in any chat interface
2. Share your source code (and any design docs, README, or partial specs)
3. The model extracts the spec structure and asks three questions:
   - Is the detected deployment type correct?
   - Should the language stay the same, or change?
   - What do you want to change/add/fix?
4. After gap-fill, it writes the complete spec
5. A `## DELTA` section captures requested changes
6. A `## MILESTONE` chain is proposed if the component is large

The output is a first-class PCD spec — not marked as `enhance-existing`.
Run `pcd-lint` on it, then translate normally.

---

## 9. Prompts Reference

| Prompt | Purpose |
|---|---|
| `prompts/interview-prompt.md` | New component: guided interview → spec |
| `prompts/reverse-prompt.md` | Existing code: reverse-engineer → spec |
| `prompts/prompt.md` | Translate spec → code (universal translator) |

All three prompts are also available via `mcp-server-pcd` as MCP resources:
- `pcd://prompts/interview`
- `pcd://prompts/reverse`
- `pcd://prompts/translator`

---

## 10. Validating Your Specification

```bash
# Basic validation
pcd-lint myspec.md

# Strict mode — warnings become errors
pcd-lint strict=true myspec.md

# List all known deployment templates
pcd-lint list-templates
```

**Exit codes:**
- `0` — valid (no errors; no warnings in strict mode)
- `1` — invalid (errors present, or warnings in strict mode)
- `2` — invocation error (bad arguments, file not found)

**Diagnostic format:**
```
ERROR  myspec.md:1  [structure]  Missing required section: ## INVARIANTS
WARNING myspec.md:6  [META]      META field 'Target' is deprecated since v0.3.0
```

**Common errors and fixes:**

| Error | Fix |
|---|---|
| Missing required section | Add the missing section |
| BEHAVIOR missing STEPS: | Add a numbered STEPS: block |
| BEHAVIOR has error exits but no negative-path EXAMPLE | Add an EXAMPLE whose THEN: shows an error outcome |
| INVARIANT missing tag | Prefix with `- [observable]` or `- [implementation]` |
| License not valid SPDX | Check https://spdx.org/licenses/ |
| MILESTONE lists unknown BEHAVIOR | The BEHAVIOR name must match exactly what is declared in the spec |
| More than one active MILESTONE | Set all but one to `pending`, `failed`, or `released` |
| Scaffold milestone not first | Move the `Scaffold: true` milestone to appear first in the file |

---

## Quick Reference — Spec Schema

```
Required sections:   META, TYPES, BEHAVIOR, PRECONDITIONS,
                     POSTCONDITIONS, INVARIANTS, EXAMPLES

Optional sections:   INTERFACES, DEPENDENCIES, TOOLCHAIN-CONSTRAINTS,
                     DELIVERABLES, MILESTONE, DELTA

BEHAVIOR variants:   ## BEHAVIOR: {name}
                     ## BEHAVIOR/INTERNAL: {name}

BEHAVIOR fields:     INPUTS, PRECONDITIONS, STEPS, POSTCONDITIONS, ERRORS
                     Constraint: required | supported | forbidden

MILESTONE fields:    Status: pending | active | failed | released
                     Scaffold: true | false  (default: false)
                     Hints-file: {comma-separated filenames}
                     Included BEHAVIORs: {names}
                     Deferred BEHAVIORs: {names}
                     Acceptance criteria: {shell commands}

INVARIANT tags:      [observable] | [implementation]

Current Spec-Schema: 0.3.21
```
