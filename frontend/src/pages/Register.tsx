
import { useState, type FormEvent } from "react";
import { registerUser } from "../services/register";

interface RegisterProps {
  onBack: () => void;
  onSuccess: () => void;
}

function Register({ onBack, onSuccess }: RegisterProps) {
  const [username, setUsername] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");

    const validUsername =
      /^(?!.*\.\.)[a-z0-9_.]{2,32}$/.test(username);

    if (!validUsername) {
      setError(
        "Use de 2 a 32 caracteres minúsculos, números, _ ou ., sem pontos consecutivos."
      );
      return;
    }

    if (
      displayName &&
      (
        Array.from(displayName).length < 2 ||
        Array.from(displayName).length > 32 ||
        displayName !== displayName.trim()
      )
    ) {
      setError("O nome deve ter entre 2 e 32 caracteres, sem espaços nas extremidades.");
      return;
    }

    if (
      Array.from(password).length < 8 ||
      Array.from(password).length > 128 ||
      !/[a-z]/.test(password) ||
      !/[A-Z]/.test(password) ||
      !/[0-9]/.test(password) ||
      !/[^\p{L}\p{N}\s]/u.test(password)
    ) {
      setError(
        "A senha deve ter de 8 a 128 caracteres, incluindo letra maiúscula, minúscula, número e símbolo."
      );
      return;
    }

    if (password !== confirmPassword) {
      setError("As senhas não coincidem.");
      return;
    }

    setLoading(true);

    try {
      await registerUser({
        username,
        display_name: displayName || null,
        password,
      });

      onSuccess();
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Erro inesperado ao criar a conta."
      );
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="login-card register-card">
      <div className="form-heading">
        <span className="welcome-label">COMECE AGORA</span>
        <h2>Crie sua conta</h2>
        <p>
          Seu espaço para criar, organizar e compartilhar links.
        </p>
      </div>

      <form onSubmit={handleSubmit}>
        <div className="field-group">
          <label htmlFor="register-username">
            Nome de usuário
          </label>
          <div className="input-wrapper">
            <span className="input-icon">@</span>
            <input
              id="register-username"
              type="text"
              placeholder="seu.usuario"
              autoComplete="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              maxLength={32}
              required
            />
          </div>
        </div>

        <div className="field-group">
          <label htmlFor="display-name">
            Nome de exibição
            <span className="optional-label"> (opcional)</span>
          </label>
          <div className="input-wrapper">
            <span className="input-icon">◇</span>
            <input
              id="display-name"
              type="text"
              placeholder="Como devemos te chamar?"
              autoComplete="nickname"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              maxLength={32}
            />
          </div>
        </div>

        <div className="field-group">
          <label htmlFor="register-password">Senha</label>
          <div className="input-wrapper">
            <span className="input-icon">⌑</span>
            <input
              id="register-password"
              type={showPassword ? "text" : "password"}
              placeholder="Crie uma senha forte"
              autoComplete="new-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
            <button
              type="button"
              className="password-toggle"
              onClick={() => setShowPassword(!showPassword)}
            >
              {showPassword ? "Ocultar" : "Mostrar"}
            </button>
          </div>
          <span className="field-hint">
            Mínimo 8 caracteres, com maiúscula, número e símbolo.
          </span>
        </div>

        <div className="field-group">
          <label htmlFor="confirm-password">Confirmar senha</label>
          <div className="input-wrapper">
            <span className="input-icon">✓</span>
            <input
              id="confirm-password"
              type={showPassword ? "text" : "password"}
              placeholder="Repita sua senha"
              autoComplete="new-password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
            />
          </div>
        </div>

        {error && (
          <div className="error-message" role="alert">
            {error}
          </div>
        )}

        <button
          type="submit"
          className="submit-button"
          disabled={loading}
        >
          {loading ? "Criando conta..." : "Criar minha conta"}
          {!loading && <span>→</span>}
        </button>
      </form>

      <div className="register-section">
        Já possui uma conta?
        <button
          type="button"
          className="auth-link"
          onClick={onBack}
        >
          Fazer login
        </button>
      </div>
    </div>
  );
}

export default Register;
