# Contributing to xparse-cli Skill

This skill wraps the `xparse-cli` CLI binary. Before reporting an issue, figure out whether the problem is in the **skill documentation** or the **CLI itself**.

## Where to report

| Problem | Where |
|---------|-------|
| Skill docs unclear, wrong example, missing flag | This repo → `skills/` |
| CLI crash, wrong output, new feature request | [xparser-ecosystem](https://github.com/textin/xparser-ecosystem/tree/main/cli) |

## Before opening an issue

1. Build or download the latest CLI:

   ```bash
   # From source
   cd cli && go build -o xparse-cli .

   # Or install via script
   source <(curl -fsSL https://dllf.intsig.net/download/2026/Solution/xparse-cli/install.sh)
   ```

2. Reproduce in your terminal — this isolates skill vs. CLI:

   ```bash
   xparse-cli auth --show          # credentials OK?
   ```

3. Note the **exit code** (`echo $?`) and the two-line stderr output (error + `> [tag] suggestion`).

## Issue template

```
**Command:** xparse-cli parse report.pdf --output ./out/
**Exit code:** 3
**Expected:** Markdown output in ./out/report.md
**Actual:** [what happened]
**CLI version:** [xparse-cli version]
**OS:** [e.g. Ubuntu 22.04, macOS 14]
```

## Updating the skill

When upstream CLI adds commands or flags:

- Add them to the correct section in SKILL.md
- Include at least one usage example per new flag
- Update the exit code table if new codes are added
- Update the supported formats table if new formats are added
- Keep the troubleshooting table current
