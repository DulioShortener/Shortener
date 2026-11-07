
const API_URL = import.meta.env.VITE_API_URL;

export interface Review {
  id: string;
  author_id: string;
  parent_id?: string;
  body: string | null;
  rating?: number;
  created_at: string;
  has_replies?: boolean;
  replies?: ReviewPage;
}

export interface ReviewPage {
  items: Review[];
  has_more: boolean;
  next_cursor: string | null;
}

interface ApiError {
  errors?: {
    code?: string;
    message?: string;
    fields?: Record<string, string>;
  };
}

class ReviewApiError extends Error {
  status: number;
  code: string;

  constructor(
    status: number,
    code: string,
    message: string
  ) {
    super(message);

    this.name = "ReviewApiError";
    this.status = status;
    this.code = code;
  }
}

async function request<T>(
  path: string,
  token?: string,
  options: RequestInit = {}
): Promise<T> {
  if (!API_URL) {
    throw new Error("A URL da API não está configurada.");
  }

  const headers = new Headers(options.headers);
  headers.set("Accept", "application/json");

  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  if (options.body) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
  });

  const text = await response.text();
  let data: unknown = null;

  if (text.trim()) {
    try {
      data = JSON.parse(text);
    } catch {
      throw new Error("O servidor retornou uma resposta inválida.");
    }
  }

  if (!response.ok) {
    const error = data as ApiError | null;

    throw new ReviewApiError(
      response.status,
      error?.errors?.code ?? "UNKNOWN_ERROR",
      error?.errors?.message ||
        Object.values(error?.errors?.fields ?? {}).join(", ") ||
        `Erro HTTP ${response.status}`
    );
  }

  return data as T;
}

// LISTAR AVALIAÇÕES
export function getReviews(
  token: string,
  cursor?: string
): Promise<ReviewPage> {
  const query = cursor
    ? `?cursor=${encodeURIComponent(cursor)}`
    : "";

  return request<ReviewPage>(`/reviews${query}`, token);
}

// BUSCAR AVALIAÇÃO DO USUÁRIO
export async function getMyReview(
  token: string
): Promise<Review | null> {
  try {
    return await request<Review>("/reviews/me", token);
  } catch (error) {
    if (
      error instanceof ReviewApiError &&
      error.status === 404 &&
      error.code === "REVIEW_NOT_FOUND"
    ) {
      return null;
    }
    throw error;
  }
}

// PUBLICAR UMA AVALIAÇÃO
export function createReview(
  token: string,
  body: string,
  rating: number
): Promise<Review> {
  return request<Review>("/reviews", token, {
    method: "POST",
    body: JSON.stringify({ body, rating }),
  });
}

// LISTAR RESPOSTAS DE UMA AVALIAÇÃO OU COMENTÁRIO
export function getReplies(
  token: string,
  parentId: string,
  cursor?: string
): Promise<ReviewPage> {
  const query = cursor
    ? `?cursor=${encodeURIComponent(cursor)}`
    : "";

  return request<ReviewPage>(
    `/reviews/${encodeURIComponent(parentId)}/replies${query}`,
    token
  );
}

// RESPONDER A UMA AVALIAÇÃO OU OUTRA RESPOSTA
export function createReply(
  token: string,
  parentId: string,
  body: string
): Promise<Review> {
  return request<Review>(
    `/reviews/${encodeURIComponent(parentId)}/replies`,
    token,
    {
      method: "POST",
      body: JSON.stringify({ body }),
    }
  );
}

// EXCLUIR AVALIAÇÃO OU RESPOSTA
export function deleteReview(
  token: string,
  id: string
): Promise<void> {
  return request<void>(
    `/reviews/${encodeURIComponent(id)}`,
    token,
    { method: "DELETE" }
  );
}
