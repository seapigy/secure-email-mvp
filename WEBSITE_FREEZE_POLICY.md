# Website Freeze Policy

## Overview

The `website/` directory is frozen to prevent unauthorized changes to the SecureMail marketing site. This policy ensures that all website modifications go through proper review, testing, and approval processes.

## Protection Mechanisms

### 1. GitHub Branch Protection
- **Required Status Check:** `website-freeze/check`
- **Required Reviewers:** `@seapigy/website-owners`
- **Dismiss Stale Reviews:** Yes
- **Require Up-to-Date Branches:** Yes

### 2. CODEOWNERS Enforcement
```
website/ @seapigy/website-owners
```

### 3. Automated Workflow Checks
- **Workflow:** `.github/workflows/website-freeze-check.yml`
- **Trigger:** PRs and pushes affecting `website/**`
- **Required Label:** `website-edit-approved`

### 4. Local Git Hooks
- **Pre-push Hook:** `.githooks/pre-push`
- **Installation:** `scripts/install-githooks.sh`
- **Warning:** Blocks pushes with website changes unless properly approved

## Change Request Process

### Step 1: Create Issue
Use the template: `.github/ISSUE_TEMPLATE/website-change-request.md`

**Required Information:**
- Purpose and justification
- Detailed description of changes
- Preview link or screenshots
- Testing steps performed
- Impact assessment
- Rollback plan

### Step 2: Get Approval
- **Required Approvers:** `@seapigy/website-owners`
- **Review Criteria:**
  - Business justification
  - Technical implementation
  - Security implications
  - Performance impact
  - User experience

### Step 3: Apply Label
- **Label Required:** `website-edit-approved`
- **Applied By:** Authorized approver only
- **Verification:** All tests must pass

### Step 4: Merge Process
- **Status Check:** `website-freeze/check` must pass
- **Visual Verification:** Manual review of changes
- **Testing:** All automated tests pass
- **Documentation:** Update relevant docs if needed

## Testing Requirements

### Automated Tests
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Build process succeeds
- [ ] Performance benchmarks met

### Manual Testing
- [ ] Visual regression testing
- [ ] Cross-browser compatibility
- [ ] Mobile responsiveness
- [ ] Accessibility compliance
- [ ] SEO impact assessment

### Performance Verification
- [ ] Bundle size within limits
- [ ] Core Web Vitals maintained
- [ ] Loading time acceptable
- [ ] No regression in metrics

## Emergency Procedures

### Critical Security Issues
1. Create issue with `[SECURITY]` prefix
2. Tag `@seapigy/website-owners` immediately
3. Provide detailed security impact
4. Follow expedited approval process
5. Document incident response

### Production Issues
1. Create issue with `[PRODUCTION]` prefix
2. Provide impact assessment
3. Include rollback plan
4. Get expedited approval
5. Monitor post-deployment

## Unfreeze Process

### Temporary Unfreeze
1. Create issue with `[UNFREEZE]` prefix
2. Specify duration and scope
3. Get approval from `@seapigy/website-owners`
4. Apply `website-edit-approved` label
5. Monitor all changes during unfreeze period
6. Re-freeze after specified duration

### Permanent Unfreeze
1. Create issue with `[PERMANENT-UNFREEZE]` prefix
2. Provide business justification
3. Get approval from repository administrators
4. Update this policy document
5. Remove protection mechanisms
6. Document decision and rationale

## Monitoring and Compliance

### Regular Audits
- Monthly review of website changes
- Quarterly policy effectiveness assessment
- Annual security review
- Performance impact analysis

### Violation Handling
- **Accidental Changes:** Revert and educate
- **Unauthorized Changes:** Immediate rollback and investigation
- **Policy Violations:** Review and update procedures
- **Security Breaches:** Incident response protocol

## Contact Information

- **Website Owners:** `@seapigy/website-owners`
- **Repository Admins:** `@seapigy`
- **Security Team:** Contact through GitHub issues
- **Emergency Contact:** Use `[URGENT]` prefix in issues

## Policy Updates

This policy can only be updated by:
1. Repository administrators
2. `@seapigy/website-owners` team
3. Following the same approval process for website changes

---

**Policy Version:** 1.0  
**Last Updated:** September 5, 2025  
**Next Review:** December 5, 2025  
**Effective Date:** September 5, 2025
