# CI Pipeline — PR Review (GLM via OpenAI SDK)

Every pull request runs two jobs defined in [`.github/workflows/pr-review.yml`](../.github/workflows/pr-review.yml):

| Job | Required check | Purpose |
|---|---|---|
| **Lint / Build / Test** | `Lint / Build / Test` | Deterministic gates: `go vet` / `go build` / `go test` (backend), `npm run build` (frontend, tsc strict). Steps self-skip until `backend/go.mod` / `frontend/package.json` exist. |
| **AI Review (GLM)** | `AI Review (GLM)` | A Python agent calls a **GLM** model via the OpenAI SDK (through Z.ai) and posts a **5W1H + Tradeoff + Suggestion** summary. |

## How it works

A small Python script (`pr_review.py`) uses the official **`openai` SDK** pointed at Z.ai (Zhipu's international GLM API), which is OpenAI-compatible. No proxy, no extra runtime:

```python
client = OpenAI(api_key=GLM_API_KEY, base_url="https://api.z.ai/api/paas/v4")
resp = client.chat.completions.create(model=GLM_MODEL, messages=[...],
                                      response_format={"type": "json_object"})
```

```text
ai-review job
  ├─ gh pr diff / view                 gather context
  ├─ openai SDK -> Z.ai/GLM (JSON)     5W1H/Tradeoff/Suggestion review
  ├─ apply INFO/MINOR search/replace   uniqueness-checked -> commit + push
  ├─ gh pr comment                     structured summary
  ├─ gh pr review --request-changes    MAJOR/CRITICAL
  └─ $GITHUB_OUTPUT has_blocker        workflow gate fails the check
```

## What the agent does

1. **Reviews** the PR diff and posts one structured comment — each finding has 5W1H (Who/What/When/Where/Why/How), the Tradeoff, and a concrete Suggestion.
2. **Auto-fixes** `INFO` / `MINOR` findings that are safe and mechanical (typos, formatting, missing error wrap) — commits the fix to the PR branch and pushes.
3. **Escalates** `MAJOR` / `CRITICAL` findings to a human: files a *request-changes* review and **fails the `AI Review (GLM)` check**, blocking merge until a human resolves it.

| Severity | Auto-fix? | Request changes? | Fails check? |
|---|---|---|---|
| INFO | ✅ (if safe) | ❌ | ❌ |
| MINOR | ✅ (if safe) | ❌ | ❌ |
| MAJOR | ❌ | ✅ | ✅ |
| CRITICAL | ❌ | ✅ | ✅ |

## Setup (one-time)

1. **Add your Z.ai (GLM) API key.** Repo → Settings → Secrets and variables → Actions → New repository secret:
   - Name: `GLM_API_KEY`
   - Value: your API key from the [Z.ai console](https://z.ai/) (used directly as a Bearer token — no JWT). If your key is from bigmodel.cn instead, change the `base_url` in `llm_client.py` to `https://open.bigmodel.cn/api/paas/v4`.
2. **(Optional) pin the GLM model.** Repo → Settings → Secrets and variables → Actions → **Variables** tab → New variable:
   - Name: `GLM_MODEL`
   - Value: a model id from Z.ai (no `z-ai/` prefix). Defaults to `glm-4.6`. Confirmed on Z.ai: `glm-4.6`, `glm-4.5`; check the console for the latest (e.g. any GLM-5.x id).
3. **Require the checks.** Repo → Settings → Branches → branch protection rule for `main` → "Require status checks to pass before merging" → add `Lint / Build / Test` and `AI Review (GLM)`.

## Files

```text
.github/
├── workflows/pr-review.yml   # pipeline (build-test + ai-review jobs)
└── agent/
    ├── pr_review.py          # orchestrator: review -> fix -> comment -> escalate
    ├── llm_client.py         # openai SDK call to Z.ai/GLM (JSON mode + fallback)
    ├── github_client.py      # gh CLI wrappers (diff, comment, review, commit)
    ├── fix_applier.py        # search/replace blocks, uniqueness-checked
    ├── review_prompt.md      # 5W1H/Tradeoff/Suggestion rubric + ShieldVN standards
    └── requirements.txt      # openai>=1.50
```

## Notes

- **Fail-open:** if the review can't complete (Z.ai/GLM outage), it posts a "skipped" comment and the check passes — a transient API issue won't block development. The deterministic build-test job remains the hard compile gate.
- **No infinite loop:** auto-fix commits use the default workflow token, so they don't re-trigger this workflow.
- To re-review after changes: push a new commit (triggers `synchronize`) or run the workflow manually with a `pr_number` input.
- The rubric enforces ShieldVN's privacy-first rule: any PII (phone, account, CCCD) reaching logs/Gemini/Sheets is flagged **CRITICAL**.
