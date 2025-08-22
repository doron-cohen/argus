export function apiFetch<T>(input: string | URL, init?: RequestInit): T {
  const path = typeof input === "string" ? input : input.toString();
  const apiHost = (globalThis as any).__ARGUS_API_HOST as string | undefined;
  const basePrefix = `${
    apiHost && apiHost.length > 0 ? apiHost.replace(/\/$/, "") : ""
  }/api/catalog/v1`;
  const base = basePrefix.startsWith("/api/")
    ? basePrefix
    : basePrefix || "/api/catalog/v1";
  const url = path.startsWith("http")
    ? path
    : `${base}${path.startsWith("/") ? "" : "/"}${path}`;

  const promise = (async () => {
    const res = await fetch(url, init);
    const data = await res.json().catch(() => ({}));
    return { status: res.status, data } as unknown as T extends Promise<infer R>
      ? R
      : never;
  })();

  return promise as unknown as T;
}
