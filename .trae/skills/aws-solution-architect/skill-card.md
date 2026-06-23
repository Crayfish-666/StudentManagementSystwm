## Description: <br>
Design AWS architectures for startups using serverless patterns and IaC templates. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[alirezarezvani](https://clawhub.ai/user/alirezarezvani) <br>

### License/Terms of Use: <br>
MIT-0 <br>


## Use Case: <br>
Developers and cloud engineers use this skill to turn startup application requirements into AWS architecture recommendations, IaC templates, deployment commands, and cost optimization guidance. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The skill includes live AWS deployment, deletion, termination, release, and cost-optimization actions without enough confirmation or safety guardrails. <br>
Mitigation: Require explicit human approval before running AWS-changing commands; prefer a sandbox account first and verify AWS account ID, region, stack name, change sets, backups or snapshots, and rollback plans before applying commands. <br>
Risk: Generated architecture and cost recommendations may be incorrect or unsuitable for a specific workload, compliance requirement, or budget. <br>
Mitigation: Review generated designs, IaC templates, IAM permissions, cost estimates, and service selections against the team's operational maturity and compliance requirements before deployment. <br>


## Reference(s): <br>
- [AWS Architecture Patterns for Startups](references/architecture_patterns.md) <br>
- [AWS Best Practices for Startups](references/best_practices.md) <br>
- [AWS Service Selection Guide](references/service_selection.md) <br>


## Skill Output: <br>
**Output Type(s):** [text, markdown, code, shell commands, configuration, guidance] <br>
**Output Format:** [Markdown guidance with JSON examples, shell commands, and IaC code snippets/templates] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Produces architecture recommendations, CloudFormation/SAM YAML, CDK TypeScript/Terraform examples, deployment checks, and cost optimization recommendations.] <br>

## Skill Version(s): <br>
2.1.1 (source: server release metadata) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
