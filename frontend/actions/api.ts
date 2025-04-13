import { apiRequest, ApiRequestOptions } from "./api-client";

export class ApiClient {
  static async fetchData<ResponseData = any, RequestBody = any>(
    path: string,
    options?: ApiRequestOptions<RequestBody>
  ) {
    return apiRequest<ResponseData, RequestBody>(path, options);
  }

  static async GET<T = any>(
    path: string,
    options?: Omit<ApiRequestOptions, "method" | "body">
  ) {
    return apiRequest<T>(path, { ...options, method: "GET" });
  }

  static async POST<ResponseData = any, RequestBody = any>(
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

  static async PUT<ResponseData = any, RequestBody = any>(
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

  static async PATCH<ResponseData = any, RequestBody = any>(
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

  static async DELETE<ResponseData = any>(
    path: string,
    options?: Omit<ApiRequestOptions, "method">
  ) {
    return apiRequest<ResponseData>(path, { ...options, method: "DELETE" });
  }
}
