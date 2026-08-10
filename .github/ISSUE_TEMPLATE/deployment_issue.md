---
name: Deployment Issue
about: Report a production deployment or release problem
title: "[Deploy] "
labels: deployment
assignees: ""
---

## Service

- [ ] API
- [ ] Analysis
- [ ] Web

## Release

- Service:
- Version tag:
- GitHub release:
- Deployment workflow:

## Deployment Result

- [ ] Deployment succeeded but service is unhealthy
- [ ] Deployment failed
- [ ] Rollback occurred
- [ ] Rollback failed
- [ ] Artifact download failed
- [ ] Artifact checksum failed
- [ ] SSH connection failed
- [ ] Service restart failed
- [ ] Other

## Failure Stage

<!-- If known, identify the deployment stage. -->

```text
VALIDATE
DOWNLOAD
VERIFY
LOCK
RESOLVE
EXTRACT
ACTIVATE
RESTART
VALIDATION
RECOVERY
PRUNE
````

## Error / Logs

```text
Paste relevant GitHub Actions or deployment logs here.

Do not include secrets, private keys, tokens, passwords, or other credentials.
```

## Expected Behavior

<!-- What should have happened? -->

## Actual Behavior

<!-- What happened instead? -->

## Recent Changes

<!-- Mention the relevant commit, PR, release, configuration, or infrastructure change. -->

## Rollback Status

* [ ] Not applicable
* [ ] Rollback completed successfully
* [ ] Rollback completed but health validation failed
* [ ] Rollback failed
* [ ] Unknown

## Additional Context

<!-- Include any other information useful for investigating the deployment. -->