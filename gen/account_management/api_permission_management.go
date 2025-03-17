/*
Dynatrace Account Management API

The enterprise management API for Dynatrace SaaS enables automation of operational tasks related to user access and environment lifecycle management.

API version: 1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package accountmanagement

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// PermissionManagementAPIService PermissionManagementAPI service
type PermissionManagementAPIService service

type ApiAddGroupPermissionsRequest struct {
	ctx            context.Context
	ApiService     *PermissionManagementAPIService
	accountUuid    string
	groupUuid      string
	permissionsDto *[]PermissionsDto
}

// The body of the request. Contains a list of permissions to be assigned to the group.   Existing permissions remain unaffected.
func (r ApiAddGroupPermissionsRequest) PermissionsDto(permissionsDto []PermissionsDto) ApiAddGroupPermissionsRequest {
	r.permissionsDto = &permissionsDto
	return r
}

func (r ApiAddGroupPermissionsRequest) Execute() (*http.Response, error) {
	return r.ApiService.AddGroupPermissionsExecute(r)
}

/*
AddGroupPermissions Assigns permissions to a user group. Existing permissions remain unaffected.

Consider upgrading your role-based permissions to IAM policies by following this guide. <a href="https://dt-url.net/gx03uwa">Learn how to manage policies<a>

	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param accountUuid The ID of the required account.    You can find the UUID on the **Account Management** > **Identity & access management** > **OAuth clients** page, during creation of an OAuth client.
	@param groupUuid The UUID of the required user group.
	@return ApiAddGroupPermissionsRequest

Deprecated
*/
func (a *PermissionManagementAPIService) AddGroupPermissions(ctx context.Context, accountUuid string, groupUuid string) ApiAddGroupPermissionsRequest {
	return ApiAddGroupPermissionsRequest{
		ApiService:  a,
		ctx:         ctx,
		accountUuid: accountUuid,
		groupUuid:   groupUuid,
	}
}

// Execute executes the request
// Deprecated
func (a *PermissionManagementAPIService) AddGroupPermissionsExecute(r ApiAddGroupPermissionsRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod = http.MethodPost
		localVarPostBody   interface{}
		formFiles          []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PermissionManagementAPIService.AddGroupPermissions")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/iam/v1/accounts/{accountUuid}/groups/{groupUuid}/permissions"
	localVarPath = strings.Replace(localVarPath, "{"+"accountUuid"+"}", url.PathEscape(parameterValueToString(r.accountUuid, "accountUuid")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"groupUuid"+"}", url.PathEscape(parameterValueToString(r.groupUuid, "groupUuid")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.permissionsDto == nil {
		return nil, reportError("permissionsDto is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.permissionsDto
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiGetGroupPermissionsRequest struct {
	ctx         context.Context
	ApiService  *PermissionManagementAPIService
	accountUuid string
	groupUuid   string
}

func (r ApiGetGroupPermissionsRequest) Execute() (*PermissionsGroupDto, *http.Response, error) {
	return r.ApiService.GetGroupPermissionsExecute(r)
}

/*
GetGroupPermissions Lists all permissions of a user group

Consider upgrading your role-based permissions to IAM policies by following this guide. <a href="https://dt-url.net/gx03uwa">Learn how to manage policies<a>

	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param accountUuid The ID of the required account.    You can find the UUID on the **Account Management** > **Identity & access management** > **OAuth clients** page, during creation of an OAuth client.
	@param groupUuid The UUID of the required user group.
	@return ApiGetGroupPermissionsRequest

Deprecated
*/
func (a *PermissionManagementAPIService) GetGroupPermissions(ctx context.Context, accountUuid string, groupUuid string) ApiGetGroupPermissionsRequest {
	return ApiGetGroupPermissionsRequest{
		ApiService:  a,
		ctx:         ctx,
		accountUuid: accountUuid,
		groupUuid:   groupUuid,
	}
}

// Execute executes the request
//
//	@return PermissionsGroupDto
//
// Deprecated
func (a *PermissionManagementAPIService) GetGroupPermissionsExecute(r ApiGetGroupPermissionsRequest) (*PermissionsGroupDto, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *PermissionsGroupDto
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PermissionManagementAPIService.GetGroupPermissions")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/iam/v1/accounts/{accountUuid}/groups/{groupUuid}/permissions"
	localVarPath = strings.Replace(localVarPath, "{"+"accountUuid"+"}", url.PathEscape(parameterValueToString(r.accountUuid, "accountUuid")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"groupUuid"+"}", url.PathEscape(parameterValueToString(r.groupUuid, "groupUuid")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiOverwriteGroupPermissionsRequest struct {
	ctx            context.Context
	ApiService     *PermissionManagementAPIService
	accountUuid    string
	groupUuid      string
	permissionsDto *[]PermissionsDto
}

// The body of the request. Contains a list of permissions to be assigned to the group.    Existing permissions are overwritten.
func (r ApiOverwriteGroupPermissionsRequest) PermissionsDto(permissionsDto []PermissionsDto) ApiOverwriteGroupPermissionsRequest {
	r.permissionsDto = &permissionsDto
	return r
}

func (r ApiOverwriteGroupPermissionsRequest) Execute() (*http.Response, error) {
	return r.ApiService.OverwriteGroupPermissionsExecute(r)
}

/*
OverwriteGroupPermissions Sets permissions of a user group. Existing permissions are overwritten.

Consider upgrading your role-based permissions to IAM policies by following this guide. <a href="https://dt-url.net/gx03uwa">Learn how to manage policies<a>

	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param accountUuid The ID of the required account.    You can find the UUID on the **Account Management** > **Identity & access management** > **OAuth clients** page, during creation of an OAuth client.
	@param groupUuid The UUID of the required user group.
	@return ApiOverwriteGroupPermissionsRequest

Deprecated
*/
func (a *PermissionManagementAPIService) OverwriteGroupPermissions(ctx context.Context, accountUuid string, groupUuid string) ApiOverwriteGroupPermissionsRequest {
	return ApiOverwriteGroupPermissionsRequest{
		ApiService:  a,
		ctx:         ctx,
		accountUuid: accountUuid,
		groupUuid:   groupUuid,
	}
}

// Execute executes the request
// Deprecated
func (a *PermissionManagementAPIService) OverwriteGroupPermissionsExecute(r ApiOverwriteGroupPermissionsRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod = http.MethodPut
		localVarPostBody   interface{}
		formFiles          []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PermissionManagementAPIService.OverwriteGroupPermissions")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/iam/v1/accounts/{accountUuid}/groups/{groupUuid}/permissions"
	localVarPath = strings.Replace(localVarPath, "{"+"accountUuid"+"}", url.PathEscape(parameterValueToString(r.accountUuid, "accountUuid")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"groupUuid"+"}", url.PathEscape(parameterValueToString(r.groupUuid, "groupUuid")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.permissionsDto == nil {
		return nil, reportError("permissionsDto is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.permissionsDto
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiRemoveGroupPermissionsRequest struct {
	ctx            context.Context
	ApiService     *PermissionManagementAPIService
	accountUuid    string
	groupUuid      string
	scope          *string
	permissionName *string
	scopeType      *string
}

// The scope of the permission to be deleted. Depending on the type of the scope, specify one of the following:    * &#x60;account&#x60;: The UUID of the account.  * &#x60;tenant&#x60;: The ID of the environment.  * &#x60;management-zone&#x60;: The ID of the management zone from an environment in &#x60;{environment-id}:{management-zone-id}&#x60; format.
func (r ApiRemoveGroupPermissionsRequest) Scope(scope string) ApiRemoveGroupPermissionsRequest {
	r.scope = &scope
	return r
}

// The name of the permission to be deleted.
func (r ApiRemoveGroupPermissionsRequest) PermissionName(permissionName string) ApiRemoveGroupPermissionsRequest {
	r.permissionName = &permissionName
	return r
}

// The scope type of the permission to be deleted.
func (r ApiRemoveGroupPermissionsRequest) ScopeType(scopeType string) ApiRemoveGroupPermissionsRequest {
	r.scopeType = &scopeType
	return r
}

func (r ApiRemoveGroupPermissionsRequest) Execute() (*http.Response, error) {
	return r.ApiService.RemoveGroupPermissionsExecute(r)
}

/*
RemoveGroupPermissions Removes a permission from a user group

Consider upgrading your role-based permissions to IAM policies by following this guide. <a href="https://dt-url.net/gx03uwa">Learn how to manage policies<a>

	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param accountUuid The ID of the required account.    You can find the UUID on the **Account Management** > **Identity & access management** > **OAuth clients** page, during creation of an OAuth client.
	@param groupUuid The UUID of the required user group.
	@return ApiRemoveGroupPermissionsRequest

Deprecated
*/
func (a *PermissionManagementAPIService) RemoveGroupPermissions(ctx context.Context, accountUuid string, groupUuid string) ApiRemoveGroupPermissionsRequest {
	return ApiRemoveGroupPermissionsRequest{
		ApiService:  a,
		ctx:         ctx,
		accountUuid: accountUuid,
		groupUuid:   groupUuid,
	}
}

// Execute executes the request
// Deprecated
func (a *PermissionManagementAPIService) RemoveGroupPermissionsExecute(r ApiRemoveGroupPermissionsRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod = http.MethodDelete
		localVarPostBody   interface{}
		formFiles          []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PermissionManagementAPIService.RemoveGroupPermissions")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/iam/v1/accounts/{accountUuid}/groups/{groupUuid}/permissions"
	localVarPath = strings.Replace(localVarPath, "{"+"accountUuid"+"}", url.PathEscape(parameterValueToString(r.accountUuid, "accountUuid")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"groupUuid"+"}", url.PathEscape(parameterValueToString(r.groupUuid, "groupUuid")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.scope == nil {
		return nil, reportError("scope is required and must be specified")
	}
	if r.permissionName == nil {
		return nil, reportError("permissionName is required and must be specified")
	}
	if r.scopeType == nil {
		return nil, reportError("scopeType is required and must be specified")
	}

	parameterAddToHeaderOrQuery(localVarQueryParams, "scope", r.scope, "form", "")
	parameterAddToHeaderOrQuery(localVarQueryParams, "permission-name", r.permissionName, "form", "")
	parameterAddToHeaderOrQuery(localVarQueryParams, "scope-type", r.scopeType, "form", "")
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}
