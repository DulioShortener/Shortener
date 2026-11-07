
import { useEffect, useState, type FormEvent } from "react";
import { login, type LoginResponse } from "./services/auth";
import {
  saveSession,
  getSavedSession,
  clearSession,
  validateSession,
  logout,
} from "./services/session";
import Register from "./pages/Register";
import Dashboard from "./pages/Dashboard";
import Reviews from "./pages/Reviews";
import "./App.css";

function LinkIcon({ size = 24 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
      <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0-7.07 7.07l1.71-1.71" />
    </svg>
  );
}

function App() {
  // ESTADOS DO LOGIN
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // SESSÃO DO USUÁRIO
  const [session, setSession] = useState<LoginResponse | null>(null);
  const [restoringSession, setRestoringSession] = useState(true);

  // NAVEGAÇÃO
  const [page, setPage] = useState<"login" | "register">("login");
  const [registerSuccess, setRegisterSuccess] = useState(false);
  const [authenticatedPage, setAuthenticatedPage] = useState<"dashboard" | "reviews">("dashboard");

  // RECUPERAR A SESSÃO AO ABRIR A PÁGINA
  useEffect(() => {
    let active = true;

    async function restoreSession() {
      const saved = getSavedSession();

      if (!saved) {
        if (active) {
          setRestoringSession(false);
        }
        return;
      }

      try {
        const user = await validateSession(saved.token);

        if (active) {
          const restoredSession = {
            ...saved,
            user,
          };

          if (Date.parse(restoredSession.expires_at) <= Date.now()) {
            clearSession();
          } else {
            setSession(restoredSession);
          }
        }
      } catch {
        if (active) {
          clearSession();
        }
      } finally {
        if (active) {
          setRestoringSession(false);
        }
      }
    }

    void restoreSession();

    return () => {
      active = false;
    };
  }, []);

  // ENCERRAR SESSÃO QUANDO O TOKEN EXPIRAR
  useEffect(() => {
    if (!session) return;

    const remaining =
      Date.parse(session.expires_at) - Date.now();

    if (!Number.isFinite(remaining) || remaining <= 0) {
      clearSession();
      setSession(null);
      return;
    }

    const timer = window.setTimeout(() => {
      clearSession();
      setSession(null);
      setUsername("");
      setPassword("");
      setPage("login");
    }, Math.min(remaining, 2147483647));

    return () => window.clearTimeout(timer);
  }, [session]);

  // REALIZAR LOGIN
  async function handleLogin(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    setError("");
    setLoading(true);

    try {
      const result = await login(username, password);

      // SALVAR TOKEN E DADOS DO USUÁRIO
      saveSession(result);

      // ABRIR DASHBOARD
      setSession(result);
      setAuthenticatedPage("dashboard");
      setRegisterSuccess(false);
      setPassword("");
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Erro inesperado ao realizar login."
      );
    } finally {
      setLoading(false);
    }
  }

  // CADASTRO REALIZADO
  function handleRegisterSuccess() {
    setRegisterSuccess(true);
    setPage("login");
    setError("");
    setPassword("");
  }

  // ABRIR CADASTRO
  function handleGoToRegister() {
    setRegisterSuccess(false);
    setError("");
    setPage("register");
  }

  // VOLTAR PARA LOGIN
  function handleGoToLogin() {
    setError("");
    setPage("login");
  }

  // SAIR DA CONTA
  async function handleLogout() {
    const token = session?.token;

    // LIMPAR SESSÃO LOCAL
    clearSession();
    setSession(null);
    setAuthenticatedPage("dashboard");
    setUsername("");
    setPassword("");
    setError("");
    setRegisterSuccess(false);
    setPage("login");

    // REVOGAR TOKEN NO BACKEND
    if (!token) return;

    try {
      await logout(token);
    } catch {
      window.alert(
        "Você saiu deste navegador, mas não foi possível confirmar o encerramento da sessão no servidor."
      );
    }
  }

  // TELA DE CARREGAMENTO DA SESSÃO
  if (restoringSession) {
    return (
      <main
        className="auth-page"
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          padding: "24px",
        }}
      >
        <div className="success-card">
          <div className="brand-icon">
            <LinkIcon />
          </div>

          <h2>Bem-vindo ao Dulio.</h2>

          <p>Verificando sua sessão...</p>
        </div>
      </main>
    );
  }

  // NAVEGAÇÃO AUTENTICADA COM UM ÚNICO CABEÇALHO
  if (session) {
    const name = session.user.display_name || session.user.username;

    return (
      <div className="app-shell">
        <header className="dashboard-header app-header">
          <div className="dashboard-brand">
            <div className="brand-icon">
              <LinkIcon size={23} />
            </div>
            <strong>Dulio.</strong>
          </div>

          <nav className="app-nav" aria-label="Navegação principal">
            <button
              type="button"
              className={`app-nav-button ${authenticatedPage === "dashboard" ? "active" : ""}`}
              aria-current={authenticatedPage === "dashboard" ? "page" : undefined}
              onClick={() => setAuthenticatedPage("dashboard")}
            >
              Meus links
            </button>
            <button
              type="button"
              className={`app-nav-button ${authenticatedPage === "reviews" ? "active" : ""}`}
              aria-current={authenticatedPage === "reviews" ? "page" : undefined}
              onClick={() => setAuthenticatedPage("reviews")}
            >
              ★ Avaliações
            </button>
          </nav>

          <div className="dashboard-user">
            <div className="user-avatar" title={name}>
              {name.charAt(0).toUpperCase()}
            </div>
            <button
              type="button"
              className="logout-button"
              onClick={() => void handleLogout()}
            >
              Sair
            </button>
          </div>
        </header>

        <div className="app-content">
          {authenticatedPage === "reviews" ? (
            <Reviews
              session={session}
              onBack={() => setAuthenticatedPage("dashboard")}
            />
          ) : (
            <Dashboard session={session} onLogout={handleLogout} />
          )}
        </div>
      </div>
    );
  }

  return (
    <main className="auth-page">
      <div className="background-glow glow-one" />
      <div className="background-glow glow-two" />

      <div className="auth-container">

        {/* LADO ESQUERDO */}
        <section className="hero-section">

          {/* LOGO */}
          <div className="brand">
            <div className="brand-icon">
              <LinkIcon size={25} />
            </div>

            <span>
              Dulio
              <span className="brand-highlight">.</span>
            </span>
          </div>

          {/* APRESENTAÇÃO */}
          <div className="hero-content">

            <div className="hero-badge">
              <span className="status-dot" />
              SEUS LINKS, SIMPLIFICADOS
            </div>

            <h1>
              Links menores.
              <br />
              <span>Possibilidades maiores.</span>
            </h1>

            <p className="hero-description">
              Transforme URLs longas em links curtos,
              fáceis de compartilhar e gerenciar.
              Tudo em um só lugar.
            </p>

            {/* DEMONSTRAÇÃO DO ENCURTADOR */}
            <div className="url-demo">

              <div className="demo-header">
                <span className="demo-dots">
                  <i />
                  <i />
                  <i />
                </span>

                <span>Como funciona</span>
              </div>

              <div className="demo-body">

                <span className="demo-label">
                  SUA URL ORIGINAL
                </span>

                <div className="url-original">
                  <span className="url-symbol">↗</span>

                  <span className="url-value">
                    https://exemplo.com/produtos/categoria/123...
                  </span>
                </div>

                <div className="conversion">
                  <div className="conversion-line" />
                  <div className="conversion-icon">↓</div>
                  <div className="conversion-line" />
                </div>

                <span className="demo-label">
                  SEU LINK ENCURTADO
                </span>

                <div className="url-short">
                  <LinkIcon size={19} />

                  <span>dulio.link/abc123</span>

                  <span className="short-check">✓</span>
                </div>

                <div className="demo-success">
                  ✓ Simples, rápido e pronto para compartilhar
                </div>

              </div>
            </div>

          </div>

          {/* RODAPÉ ESQUERDO */}
          <div className="hero-footer">
            © {new Date().getFullYear()} Dulio Shortener
          </div>

        </section>

        {/* LADO DIREITO */}
        <section className="form-section">

          {page === "register" ? (

            /* TELA DE CADASTRO */
            <Register
              onBack={handleGoToLogin}
              onSuccess={handleRegisterSuccess}
            />

          ) : (

            /* TELA DE LOGIN */
            <div className="login-card">

              {/* LOGO MOBILE */}
              <div className="mobile-brand">
                <div className="brand-icon">
                  <LinkIcon />
                </div>

                <span>Dulio.</span>
              </div>

              {/* CABEÇALHO */}
              <div className="form-heading">
                <span className="welcome-label">
                  BEM-VINDO DE VOLTA
                </span>

                <h2>Acesse sua conta</h2>

                <p>
                  Entre com suas credenciais para continuar.
                </p>
              </div>

              {/* MENSAGEM APÓS CADASTRO */}
              {registerSuccess && (
                <div
                  className="success-message"
                  role="status"
                >
                  ✓ Conta criada com sucesso!
                  Faça login para continuar.
                </div>
              )}

              {/* FORMULÁRIO */}
              <form onSubmit={handleLogin}>

                {/* USUÁRIO */}
                <div className="field-group">
                  <label htmlFor="username">
                    Nome de usuário
                  </label>

                  <div className="input-wrapper">
                    <span className="input-icon">＠</span>

                    <input
                      id="username"
                      type="text"
                      placeholder="Digite seu usuário"
                      autoComplete="username"
                      value={username}
                      onChange={(e) =>
                        setUsername(e.target.value)
                      }
                      required
                    />
                  </div>
                </div>

                {/* SENHA */}
                <div className="field-group">
                  <label htmlFor="password">
                    Senha
                  </label>

                  <div className="input-wrapper">
                    <span className="input-icon">⌑</span>

                    <input
                      id="password"
                      type={
                        showPassword
                          ? "text"
                          : "password"
                      }
                      placeholder="Digite sua senha"
                      autoComplete="current-password"
                      value={password}
                      onChange={(e) =>
                        setPassword(e.target.value)
                      }
                      required
                    />

                    <button
                      type="button"
                      className="password-toggle"
                      onClick={() =>
                        setShowPassword(!showPassword)
                      }
                      aria-label={
                        showPassword
                          ? "Ocultar senha"
                          : "Mostrar senha"
                      }
                    >
                      {showPassword
                        ? "Ocultar"
                        : "Mostrar"}
                    </button>
                  </div>
                </div>

                {/* ERRO */}
                {error && (
                  <div
                    className="error-message"
                    role="alert"
                  >
                    {error}
                  </div>
                )}

                {/* BOTÃO LOGIN */}
                <button
                  type="submit"
                  className="submit-button"
                  disabled={loading}
                >
                  {loading
                    ? "Entrando..."
                    : "Entrar na minha conta"}

                  {!loading && <span>→</span>}
                </button>

              </form>

              {/* IR PARA CADASTRO */}
              <div className="register-section">
                Ainda não tem uma conta?

                <button
                  type="button"
                  className="auth-link"
                  onClick={handleGoToRegister}
                >
                  Criar conta
                </button>
              </div>

            </div>

          )}

          {/* RODAPÉ DIREITO */}
          <div className="form-footer">
            Encurte. Compartilhe. Simplifique.
          </div>

        </section>

      </div>
    </main>
  );
}

export default App;
