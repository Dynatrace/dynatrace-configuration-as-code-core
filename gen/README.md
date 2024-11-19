# Dynatrace configuration as code core library
## Generated clients

### How to re-generate the Account Management client

1. Create a new branch
2. Update the spec:
    1. Download the latest Open API spec from `https://api.dynatrace.com/spec-json` 
    2. Save it as `gen/specs/account_management/spec.json`.
    3. Format that spec to make it readable and save it as `gen/specs/account_management/spec_formatted.json`
    4. Make any manual changes to spec as needed, for example, while waiting for official changes.

3. Generate the code by running `./gen.sh` 
4. Selectively revert a few changes in `gen/account_management/client.go` - maybe this wont be needed in the future
    1. Preserve custom implementation of `func parameterAddToHeaderOrQuery(headerOrQueryParams interface{}, keyPrefix string, obj interface{}, style string, collectionType string)`
    2. Remove duplicated definitions of `reportError(...)` and `newStrictDecoder(...)`
5. Check that everything builds and run all tests, or commit and push the code, open a draft PR and ensure all checks pass
6. Update the core-library in Monaco and the Dynatrace Terraform provider and check that all tests, including E2E tests, pass there.
