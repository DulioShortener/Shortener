
const API_URL = import.meta.env.VITE_API_URL;

export interface ShortLink {
  id: string;
  user_id: string;
  code: string;
  target_url: string;
  short_url: string;
  created_at: string;
}

interface LinksResponse {
  links: ShortLink[];
}

interface ApiError {
  errors?: {
    code?: string;
    message?: string;
    fields?: Record<string, string>;
  };
}

async function request<T>(
  endpoint: string,
  token: string,
  options: RequestInit = {}
): Promise<T> {
  if (!API_URL) {
    throw new Error("URL da API não configurada.");
  }

  const headers = new Headers(options.headers);
  headers.set("Authorization", `Bearer ${token}`);
  headers.set("Accept", "application/json");

  if (options.body) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${API_URL}${endpoint}`, {
    ...options,
    headers,
  });

  const text = await response.text();

  let data: unknown = null;

  if (text.trim()) {
    try {
      data = JSON.parse(text);
    } catch {
      throw new Error("A API retornou uma resposta inválida.");
    }
  }

  if (!response.ok) {
    const error = data as ApiError | null;

    throw new Error(
      error?.errors?.message ||
      Object.values(error?.errors?.fields ?? {}).join(", ") ||
      `Erro na requisição (HTTP ${response.status}).`
    );
  }

  return data as T;
}

export async function getLinks(token: string) {
  const result = await request<LinksResponse>("/links", token);
  return result.links;
}

export async function createLink(url: string, token: string) {
  return request<ShortLink>("/links", token, {
    method: "POST",
    body: JSON.stringify({ url }),
  });
}

export async function deleteLink(id: string, token: string) {
  return request<void>(`/links/${encodeURIComponent(id)}`, token, {
    method: "DELETE",
  });
}
