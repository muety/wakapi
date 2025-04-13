import { getSession } from "./session";

export type Result<D, E = Error> =
  | { success: true; data: D; error: null }
  | { success: false; data: null; error: E };

type HttpMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

export type ApiRequestOptions<T = any> = {
  method?: HttpMethod;
  body?: T;
  headers?: Record<string, string>;
  cache?: RequestCache;
  next?: NextFetchRequestConfig;
  skipAuth?: boolean; // New option to skip authentication
};

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function apiRequest<ResponseData = any, RequestBody = any>(
  path: string,
  options?: ApiRequestOptions<RequestBody>
): Promise<Result<ResponseData, Error>> {
  try {
    const url = `${API_URL}/api/${path}`;
    const {
      method = "GET",
      body,
      headers = {},
      cache,
      next,
      skipAuth = false,
    } = options || {};
    const requestHeaders: Record<string, string> = {
      "Content-Type": "application/json",
      Accept: "application/json",
      ...headers,
    };
    if (!skipAuth) {
      const session = await getSession();
      const token = session?.token;
      if (!token) {
        return {
          success: false,
          data: null,
          error: new Error("Authentication required"),
        };
      }
      requestHeaders["Token"] = token;
    }
    const response = await fetch(url, {
      method,
      headers: requestHeaders,
      body: body ? JSON.stringify(body) : undefined,
      cache,
      next,
    });
    // Handle non-JSON responses
    const contentType = response.headers.get("content-type");
    let data: ResponseData;
    if (contentType?.includes("application/json")) {
      data = await response.json();
    } else {
      // Handle other response types or empty responses
      data = (await response.text()) as unknown as ResponseData;
    }
    if (!response.ok) {
      return {
        success: false,
        data: null,
        error: new Error(
          `API error (${response.status}): ${response.statusText || JSON.stringify(data)}`
        ),
      };
    }
    return {
      success: true,
      data,
      error: null,
    };
  } catch (error) {
    return {
      success: false,
      data: null,
      error: error instanceof Error ? error : new Error(String(error)),
    };
  }
}
