## Description: <br>
Use when medium-to-large changes need explicit requirements, technical design, and task planning before implementation, especially for multi-module work, unclear acceptance criteria, or architecture-heavy requests. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[binggg](https://clawhub.ai/user/binggg) <br>

### License/Terms of Use: <br>
MIT-0 <br>


## Use Case: <br>
Developers and engineering agents use this skill to turn medium or large coding requests into explicit requirements, technical design notes, and implementation tasks before execution. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The workflow can slow larger coding tasks by requiring staged requirements, design, and task planning before implementation. <br>
Mitigation: Use it for medium or large changes with unclear acceptance criteria, and skip the full workflow for small, precise, low-risk requests. <br>
Risk: The skill may direct the agent to related CloudBase reference documents for UI or data-model work. <br>
Mitigation: Review those referenced documents before use in environments that restrict external documentation. <br>


## Reference(s): <br>
- [ClawHub skill page](https://clawhub.ai/binggg/spec-workflow-guide) <br>
- [CloudBase main entry](https://cnb.cool/tencent/cloud/cloudbase/cloudbase-skills/-/git/raw/main/skills/cloudbase/SKILL.md) <br>
- [Spec workflow raw source](https://cnb.cool/tencent/cloud/cloudbase/cloudbase-skills/-/git/raw/main/skills/cloudbase/references/spec-workflow/SKILL.md) <br>
- [UI design sibling reference](https://cnb.cool/tencent/cloud/cloudbase/cloudbase-skills/-/git/raw/main/skills/cloudbase/references/ui-design/SKILL.md) <br>
- [Data model sibling reference](https://cnb.cool/tencent/cloud/cloudbase/cloudbase-skills/-/git/raw/main/skills/cloudbase/references/data-model-creation/SKILL.md) <br>


## Skill Output: <br>
**Output Type(s):** [Markdown, Guidance, Configuration] <br>
**Output Format:** [Markdown documents and concise planning guidance] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [May create requirements, design, and task files under a specs directory after confirming the workflow with the user.] <br>

## Skill Version(s): <br>
1.17.0 (source: ClawHub release metadata; artifact frontmatter lists 2.21.0) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
