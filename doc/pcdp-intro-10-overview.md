# What

## Human Intent, Machine Implementation

::::columns
:::: {.column width=60%}

Domain experts write **specifications**.

AI generates **all** implementation code.

Engineers never write implementation code directly.

::::
:::: {.column width=40%}

![](pcdp-logo-green.png){height=5cm}

::::
::::

---

## This is not "AI-assisted coding"

| | Traditional | Vibe Coding | **PCDP** |
|---|---|---|---|
| Human writes | code | code + prompts | **specs only** |
| AI role | none | suggests | **translates** |
| Primary artifact | source code | source code | **specification** |
| Target language | developer | developer | **template** |
| Regulated domains | manual audit | prohibited | **enabled** |

---

## The specification

::::columns
:::: {.column width=50%}

```markdown
## BEHAVIOR: transfer
INPUTS:
  from: Account
  amount: Amount
STEPS:
  1. Validate preconditions.
  2. Debit from.balance.
  3. Credit to.balance.
POSTCONDITIONS:
  - SUM(balances) unchanged
EXAMPLES:
  EXAMPLE: success
  GIVEN: balance = 100
  WHEN:  transfer(30)
  THEN:  balance = 70
```

::::
:::: {.column width=50%}

No programming language.

No target platform.

No toolchain knowledge required.

The **template** decides all of that.

::::
::::

# Why

## The problem: AI is blocked where it matters most

Regulated software markets **cannot use** current AI code generation:

- **Automotive** — ISO 26262
- **Aviation** — DO-178C
- **Medical devices** — IEC 62304
- **Security certification** — Common Criteria / EUCC

Reason: AI-generated code is opaque. No audit trail. Regulators reject it.

\bigskip

Automotive software alone: **\$50B+ annually**.

---

## The opportunity

::::columns
:::: {.column width=55%}

**Specifications are auditable.**

Regulators review the spec — not the code.

The AI is a translator, not an author.

Verifiability comes from:

- human-reviewable specs
- formal proofs (optional)
- independent test generation
- full audit bundles

::::
:::: {.column width=45%}

**Digital sovereignty.**

Works with locally-hosted models.

No US cloud dependency.

Tested: 120B open-weight model\
at a regional EU provider\
— most complete output\
of any tested model.

::::
::::

---

## Proof: pcdp-lint

`pcdp-lint` — the specification validator — was\
**specified and generated using PCDP itself.**

Zero hand-written implementation code.

\bigskip

Tested across multiple AI models,\
three continents, one result:

\bigskip

\begin{center}
\Large Every model resolved Go from the template\\
\large without being told.
\end{center}

---

# How

## The workflow

![](pcdp-workflow.png){height=3.5cm}

\bigskip

1. Domain expert writes a spec (or AI interviews them)
2. `pcdp-lint` validates structure
3. Deployment template resolves the language
4. LLM translates spec → code + audit bundle

---

## Language is never your problem

![](pcdp-resolution.png){height=3cm}

\bigskip

The spec declares **what** and **where** (deployment context).

The template resolves **language, packaging, conventions**.

A spec written today is valid if the organisation\
switches from Go to Rust in 2029 — **no spec change needed**.

---

## The audit bundle

::::columns
:::: {.column width=50%}

Every translation produces:

- specification (human-reviewable)
- generated source code
- packaging artifacts (RPM, DEB, OCI)
- independent test suite
- translation report
- workflow diagram (Pikchr)
- metadata + hashes

::::
:::: {.column width=50%}

Designed for:

- ISO 26262 (automotive)
- DO-178C (aviation)
- IEC 62304 (medical)
- Common Criteria EAL4+
- EU Cyber Resilience Act

::::
::::

---

## Getting started

::::columns
:::: {.column width=50%}

**Write a spec (no format knowledge needed):**

```bash
# AI interviews the domain expert:
ollama run llama3.2 \
  "$(cat prompts/interview-prompt.md)"
```

Then validate:

```bash
pcdp-lint myspec.md
```

::::
:::: {.column width=50%}

**Translate to code:**

```bash
# mcphost with the translator prompt
# reads template + spec,
# produces code + audit bundle
```

\bigskip

Everything in the repo:\
**github.com/mge1512/pcdp**

::::
::::
