
const API_URL = import.meta.env.VITE_API_URL;

export interface User {
  id: string;
  username: string;
  display_name: string | null;
  created_at: string;
  updated_at: string;
}

export interface LoginResponse {
  id: string;
  token: string;
  token_type: string;
  user: User;
  created_at: string;
  expires_at: string;
}

export async function login(
  username: string,
  password: string
): Promise<LoginResponse> {
  const response = await fetch(`${API_URL}/auth/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ username, password }),
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(
      data?.errors?.message ||
      Object.values(data?.errors?.fields ?? {}).join(", ") ||
      "Não foi possível realizar o login."
    );
  }

  return data as LoginResponse;
}
