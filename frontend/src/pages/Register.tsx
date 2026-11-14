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

    const validUsername = /^(?!.*\.\.)[a-z0-9_.]{2,32}$/.test(username);
    if (!validUsername) {
      setError("Use de 2 a 32 letras minúsculas, números, _ ou ., sem pontos consecutivos.");
      return;
    }

    if (
      displayName &&
      (Array.from(displayName).length < 2 ||
        Array.from(displayName).length > 32 ||
        displayName !== displayName.trim())
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
      setError("A senha deve ter de 8 a 128 caracteres, incluindo maiúscula, minúscula, número e símbolo.");
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
      setError(err instanceof Error ? err.message : "Erro inesperado ao criar a conta.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="login-card register-card brutal-register">
      <div className="br-register-topline">
        <span className="br-register-kicker">DULIO / ACESSO</span>
        <span className="br-register-index">02 — CADASTRO</span>
      </div>

      <div className="form-heading br-register-heading">
        <span className="welcome-label">NOVA CONTA</span>
        <h2>ENTRE PARA O <span>TIME.</span></h2>
        <p>Crie sua conta para encurtar, organizar e compartilhar seus links.</p>
      </div>

      <form onSubmit={handleSubmit} noValidate={false}>
        <div className="field-group">
          <label htmlFor="register-username">01 / Nome de usuário</label>
          <div className="input-wrapper">
            <span className="input-icon" aria-hidden="true">@</span>
            <input
              id="register-username"
              type="text"
              placeholder="seu.usuario"
              autoComplete="username"
              autoCapitalize="none"
              spellCheck={false}
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              maxLength={32}
              required
            />
          </div>
          <span className="field-hint">2–32 caracteres · a–z, 0–9, _ e .</span>
        </div>

        <div className="field-group">
          <label htmlFor="display-name">
            02 / Nome de exibição <span className="optional-label">— opcional</span>
          </label>
          <div className="input-wrapper">
            <span className="input-icon" aria-hidden="true">◇</span>
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
          <label htmlFor="register-password">03 / Senha</label>
          <div className="input-wrapper">
            <span className="input-icon" aria-hidden="true">#</span>
            <input
              id="register-password"
              type={showPassword ? "text" : "password"}
              placeholder="Crie sua senha"
              autoComplete="new-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              minLength={8}
              maxLength={128}
              required
            />
            <button
              type="button"
              className="password-toggle"
              onClick={() => setShowPassword((current) => !current)}
              aria-label={showPassword ? "Ocultar senha" : "Mostrar senha"}
              aria-pressed={showPassword}
            >
              {showPassword ? "OCULTAR" : "MOSTRAR"}
            </button>
          </div>
          <span className="field-hint">8–128 caracteres · maiúscula, minúscula, número e símbolo</span>
        </div>

        <div className="field-group">
          <label htmlFor="confirm-password">04 / Confirmar senha</label>
          <div className="input-wrapper">
            <span className="input-icon" aria-hidden="true">✓</span>
            <input
              id="confirm-password"
              type={showPassword ? "text" : "password"}
              placeholder="Repita sua senha"
              autoComplete="new-password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              minLength={8}
              maxLength={128}
              required
            />
          </div>
        </div>

        {error && (
          <div className="error-message" role="alert">{error}</div>
        )}

        <button type="submit" className="submit-button br-register-submit" disabled={loading}>
          <span>{loading ? "CRIANDO CONTA..." : "CRIAR CONTA"}</span>
          {!loading && <span aria-hidden="true">↗</span>}
        </button>
      </form>

      <div className="register-section">
        <span>JÁ TEM UMA CONTA?</span>
        <button type="button" className="auth-link" onClick={onBack}>
          FAZER LOGIN →
        </button>
      </div>
    </div>
  );
}

export default Register;
