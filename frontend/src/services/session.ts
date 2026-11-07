
import type { LoginResponse, User } from "./auth";

const API_URL = import.meta.env.VITE_API_URL;
const SESSION_KEY = "dulio_session";

export function saveSession(session: LoginResponse) {
  sessionStorage.setItem(SESSION_KEY, JSON.stringify(session));
}

export function clearSession() {
  sessionStorage.removeItem(SESSION_KEY);
}

export function getSavedSession(): LoginResponse | null {
  const saved = sessionStorage.getItem(SESSION_KEY);

  if (!saved) return null;

  try {
    const session = JSON.parse(saved) as LoginResponse;

    if (
      typeof session.token !== "string" ||
      !session.token ||
      !session.user ||
      typeof session.user.username !== "string" ||
      !Number.isFinite(Date.parse(session.expires_at)) ||
      Date.parse(session.expires_at) <= Date.now()
    ) {
      clearSession();
      return null;
    }

    return session;
  } catch {
    clearSession();
    return null;
  }
}

export async function validateSession(
  token: string
): Promise<User> {
  const response = await fetch(`${API_URL}/users/me`, {
    headers: {
      Authorization: `Bearer ${token}`,
      Accept: "application/json",
    },
  });

  if (!response.ok) {
    throw new Error(
      response.status === 401
        ? "Sessão expirada ou inválida."
        : `Não foi possível verificar a sessão (${response.status}).`
    );
  }

  return response.json() as Promise<User>;
}

export async function logout(token: string): Promise<void> {
  const response = await fetch(`${API_URL}/auth/logout`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      Accept: "application/json",
    },
  });

  if (!response.ok) {
    throw new Error(
      `Não foi possível revogar a sessão (HTTP ${response.status}).`
    );
  }
}
