# Rules (P0-P1 adversarial tests)
1. All test cases must use authenticated superuser context for setup but test with non-superuser/missing tenant selector for isolation cases
2. Test must verify that cross-tenant access is rejected with appropriate 403/404 status
3. Suspended/deleting tenants must fail closed - no CRUD operations allowed
4. Error/DTO responses must never expose internal paths, DSNs, credentials, or secret references
5. Tests must be independent; no shared test state between test cases
6. Use existing test fixtures from arsy/httpapi_test package
7. All test functions must follow the pattern: test name, fixture setup, action, assertion
8. Do not modify existing test infrastructure; add new test functions only
9. Test coverage must include: wrong tenant, missing tenant, member of A accessing B, inactive membership, suspended/deleting tenant
10. Report any test failures as BLOCKED and document deviations
