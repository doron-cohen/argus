declare global {
  var __ARGUS_API_HOST: string | undefined;
}

const getUrl = (contextUrl: string): string => {
  const url = new URL(contextUrl, "http://localhost");
  const pathname = url.pathname;
  const search = url.search;
  const apiHost = globalThis.__ARGUS_API_HOST || "http://localhost:8080";

  // If apiHost is empty, use relative URLs (for production build served by backend)
  if (apiHost === "") {
    return `${pathname}${search}`;
  }

  const baseUrl = apiHost.replace(/\/$/, "");
  const requestUrl = new URL(`${baseUrl}${pathname}${search}`);
  return requestUrl.toString();
};

export const apiFetch = async <T>(
  url: string,
  options: RequestInit,
): Promise<T> => {
  const requestUrl = getUrl(url);
  const requestInit: RequestInit = {
    ...options,
  };

  const request = new Request(requestUrl, requestInit);
  const response = await fetch(request);
  const data = await response.json();

  // Handle the case where T is Promise<U> by unwrapping it
  const result = { status: response.status, data };
  return result as any;
};
