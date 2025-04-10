"use client";

import { apiRequest, ApiRequestOptions } from "./api-client";

export async function fetchData<ResponseData = any, RequestBody = any>(
  path: string,
  options?: ApiRequestOptions<RequestBody>
) {
  return apiRequest<ResponseData, RequestBody>(path, options);
}

// Specialized API functions (optional, for common operations)
export async function getData<T = any>(
  path: string,
  options?: Omit<ApiRequestOptions, "method" | "body">
) {
  return apiRequest<T>(path, { ...options, method: "GET" });
}

export async function postData<ResponseData = any, RequestBody = any>(
  path: string,
  body: RequestBody,
  options?: Omit<ApiRequestOptions, "method" | "body">
) {
  return apiRequest<ResponseData, RequestBody>(path, {
    ...options,
    method: "POST",
    body,
  });
}

export async function updateData<ResponseData = any, RequestBody = any>(
  path: string,
  body: RequestBody,
  options?: Omit<ApiRequestOptions, "method" | "body">
) {
  return apiRequest<ResponseData, RequestBody>(path, {
    ...options,
    method: "PUT",
    body,
  });
}

export async function patchData<ResponseData = any, RequestBody = any>(
  path: string,
  body: RequestBody,
  options?: Omit<ApiRequestOptions, "method" | "body">
) {
  return apiRequest<ResponseData, RequestBody>(path, {
    ...options,
    method: "PATCH",
    body,
  });
}

export async function deleteData<ResponseData = any>(
  path: string,
  options?: Omit<ApiRequestOptions, "method">
) {
  return apiRequest<ResponseData>(path, { ...options, method: "DELETE" });
}
