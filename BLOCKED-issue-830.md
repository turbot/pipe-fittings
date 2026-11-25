## Kind: design

# Issue #830: nothing left to do until a person says the security audit may run here

**What I was doing.** Reading #830 against the base. PR #831 (merged) already did the dependency
chore part: `.claude/agent-fleet/agent-fleet.json` has `dependencies` as `{ "when": "tue week 1", "mergeByHand": true }`
and `secret-scan` stays off, as the issue asks.

**What is left.** `security-audit` and `dependabot-cant-fix` still carry `off: true`. The issue asks to
remove both. The maintainer (e-gineer) reopened the issue by hand after #831 merged and wrote:
"This issue stays open, and held, for the other two chores. The security audit stays off here until it
can file findings in this public repo without spelling out how to use them. 'Dependabot can't fix'
waits with it."

**What I ruled out.** Removing the two `off` entries would turn on a public-issue security audit the
maintainer has said must stay off until a fleet capability exists (filing findings without exploit
detail). That is a decision a person took, and it is not mine to override. This is not a `waiting`
case: no PR or issue is named that delivers the capability, so there is nothing to wait on.

**Question / next step.** Once the fleet can file public security findings without exploit detail,
remove `off` from `security-audit` and `dependabot-cant-fix` in `.claude/agent-fleet/agent-fleet.json`
(delete entries left empty), and re-arm this issue. Or, if the two chores should stay off for good,
close #830 as completed by #831.
