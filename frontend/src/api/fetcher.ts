declare global {
  var __ARGUS_API_HOST: string | undefined;
}

export function apiFetch<T>(input: string | URL, init?: RequestInit): T {
  const path = typeof input === "string" ? input : input.toString();
  const apiHost = globalThis.__ARGUS_API_HOST;
  const basePrefix = `${
    apiHost && apiHost.length > 0 ? apiHost.replace(/\/$/, "") : ""
  }/api/catalog/v1`;

  const isAbsoluteUrl = /^https?:\/\//.test(basePrefix);
  const base = isAbsoluteUrl
    ? basePrefix
    : basePrefix.startsWith("/api/")
      ? basePrefix
      : `/${basePrefix}`;

  const url = `${base}${path.startsWith("/") ? path : `/${path}`}`;

  return fetch(url, init) as T;
}
