export function apiFetch<T>(input: string | URL, init?: RequestInit): T {
  const path = typeof input === "string" ? input : input.toString();
  const url = path.startsWith("http")
    ? path
    : `/api/catalog/v1${path.startsWith("/") ? "" : "/"}${path}`;

  const promise = (async () => {
    const res = await fetch(url, init);
    const data = await res.json().catch(() => ({}));
    return { status: res.status, data } as unknown as T extends Promise<infer R>
      ? R
      : never;
  })();

  return promise as unknown as T;
}
