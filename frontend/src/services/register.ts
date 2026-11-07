
import type { User } from "./auth";

const API_URL = import.meta.env.VITE_API_URL;

export interface RegisterData {
  username: string;
  display_name?: string | null;
  password: string;
}

interface ApiError {
  errors?: {
    code?: string;
    message?: string;
    fields?: Record<string, string>;
  };
}

export async function registerUser(
  userData: RegisterData
): Promise<User> {
  const response = await fetch(`${API_URL}/users`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
    },
    body: JSON.stringify(userData),
  });

  // Lê a resposta como texto primeiro
  const responseText = await response.text();

  let data: (Partial<User> & ApiError) | null = null;

  if (responseText.trim()) {
    try {
      data = JSON.parse(responseText);
    } catch {
      throw new Error(
        `A API retornou uma resposta inválida (HTTP ${response.status}).`
      );
    }
  }

  // Verifica os erros enviados pelo backend
  if (!response.ok) {
    if (data?.errors?.code === "USERNAME_TAKEN") {
      throw new Error("Esse nome de usuário já está em uso.");
    }

    if (data?.errors?.code === "RATE_LIMIT_EXCEEDED") {
      throw new Error(
        "Muitas tentativas. Tente novamente mais tarde."
      );
    }

    const fieldErrors = Object.values(
      data?.errors?.fields ?? {}
    );

    throw new Error(
      fieldErrors.join(" | ") ||
      data?.errors?.message ||
      `Erro ao criar conta (HTTP ${response.status}).`
    );
  }

  // A documentação exige HTTP 201 com dados do usuário
  if (
    response.status !== 201 ||
    !data ||
    typeof data.id !== "string" ||
    typeof data.username !== "string"
  ) {
    throw new Error(
      "A API não retornou a confirmação esperada. Verifique se a conta foi criada antes de tentar novamente."
    );
  }

  return data as User;
}
