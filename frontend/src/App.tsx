
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

  // Carregamento da sessão
  if (restoringSession) {
    return (
      <main className="db-auth db-auth-loading">
        <div className="db-loading-panel" role="status" aria-live="polite">
          <div className="db-logo-mark"><LinkIcon size={28} /></div>
          <strong>DULIO<span className="db-accent-dot">.</span></strong>
          <p>Verificando sua sessão...</p>
        </div>
      </main>
    );
  }

  // Área logada: cabeçalho compartilhado entre links e avaliações
  if (session) {
    const name = session.user.display_name || session.user.username;

    return (
      <div className="app-shell db-shell">
        <header className="dashboard-header app-header db-header">
          <div className="dashboard-brand db-brand">
            <span className="brand-icon db-logo-mark" aria-hidden="true">
              <LinkIcon size={27} />
            </span>
            <span className="db-brand-type">
              <strong>DULIO</strong>
              <small>SHORTENER</small>
            </span>
          </div>

          <nav className="app-nav db-nav" aria-label="Navegação principal">
            <button
              type="button"
              className={`app-nav-button db-nav-button ${authenticatedPage === "dashboard" ? "active" : ""}`}
              aria-current={authenticatedPage === "dashboard" ? "page" : undefined}
              onClick={() => setAuthenticatedPage("dashboard")}
            >
              LINKS
            </button>
            <button
              type="button"
              className={`app-nav-button db-nav-button ${authenticatedPage === "reviews" ? "active" : ""}`}
              aria-current={authenticatedPage === "reviews" ? "page" : undefined}
              onClick={() => setAuthenticatedPage("reviews")}
            >
              AVALIAÇÕES
            </button>
          </nav>

          <div className="dashboard-user db-user">
            <div className="user-avatar db-avatar" title={name} aria-hidden="true">
              {name.charAt(0).toUpperCase()}
            </div>
            <div className="db-user-info">
              <strong>{name}</strong>
              <span>Minha conta</span>
            </div>
            <button
              type="button"
              className="logout-button db-logout"
              onClick={() => void handleLogout()}
            >
              <span aria-hidden="true">↗</span> SAIR
            </button>
          </div>
        </header>

        <div className="app-content db-content">
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

  // Login e cadastro — visual brutalista, sem mudar o fluxo de autenticação
  return (
    <main className="auth-page db-auth">
      <div className="auth-container db-auth-layout">
        <section className="hero-section db-auth-hero">
          <div className="brand db-auth-brand">
            <span className="brand-icon db-logo-mark"><LinkIcon size={28} /></span>
            <span className="db-brand-type">
              <strong>DULIO</strong>
              <small>SHORTENER</small>
            </span>
          </div>

          <div className="hero-content db-auth-hero-content">
            <span className="db-section-index">[ 01 / SHORTENER ]</span>
            <h1>
              LINKS LONGOS<span className="db-accent-dot">.</span>
              <span className="db-hero-highlight">CAMINHOS CURTOS.</span>
            </h1>
            <p className="hero-description db-hero-description">
              Encurte seus links, copie e compartilhe. Sem complicação.
            </p>

            <div className="db-flow-demo" aria-label="Exemplo de encurtamento">
              <div className="db-flow-label">COMO FUNCIONA</div>
              <div className="db-flow-source">
                <span>01 / URL ORIGINAL</span>
                <code>https://exemplo.com/caminho/muito/longo</code>
              </div>
              <div className="db-flow-arrow" aria-hidden="true">↓ ↓ ↓</div>
              <div className="db-flow-result">
                <span>02 / LINK CURTO</span>
                <strong><LinkIcon size={19} /> dulio.link/abc123</strong>
              </div>
            </div>
          </div>

          <footer className="hero-footer db-hero-footer">
            © {new Date().getFullYear()} DULIO SHORTENER
          </footer>
        </section>

        <section className="form-section db-auth-form-section">
          <div className="db-form-topline">ACESSO / DULIO SHORTENER <span>↗</span></div>

          {page === "register" ? (
            <Register
              onBack={handleGoToLogin}
              onSuccess={handleRegisterSuccess}
            />
          ) : (
            <div className="login-card db-login-panel">
              <div className="mobile-brand db-mobile-brand">
                <span className="brand-icon db-logo-mark"><LinkIcon /></span>
                <strong>DULIO.</strong>
              </div>

              <div className="form-heading db-form-heading">
                <span className="welcome-label db-overline">ÁREA DE ACESSO</span>
                <h2>ENTRAR<span className="db-accent-dot">.</span></h2>
                <p>Use sua conta para gerenciar seus links.</p>
              </div>

              {registerSuccess && (
                <div className="success-message" role="status">
                  ✓ Conta criada com sucesso. Faça login para continuar.
                </div>
              )}

              <form onSubmit={handleLogin} className="db-login-form">
                <div className="field-group db-field">
                  <label htmlFor="username">NOME DE USUÁRIO</label>
                  <div className="input-wrapper db-input-wrap">
                    <input
                      id="username"
                      type="text"
                      placeholder="Seu usuário"
                      autoComplete="username"
                      value={username}
                      onChange={(event) => setUsername(event.target.value)}
                      required
                    />
                  </div>
                </div>

                <div className="field-group db-field">
                  <label htmlFor="password">SENHA</label>
                  <div className="input-wrapper db-input-wrap">
                    <input
                      id="password"
                      type={showPassword ? "text" : "password"}
                      placeholder="Sua senha"
                      autoComplete="current-password"
                      value={password}
                      onChange={(event) => setPassword(event.target.value)}
                      required
                    />
                    <button
                      type="button"
                      className="password-toggle db-password-toggle"
                      onClick={() => setShowPassword((current) => !current)}
                      aria-label={showPassword ? "Ocultar senha" : "Mostrar senha"}
                      aria-pressed={showPassword}
                    >
                      {showPassword ? "OCULTAR" : "MOSTRAR"}
                    </button>
                  </div>
                </div>

                {error && <div className="error-message" role="alert">{error}</div>}

                <button
                  type="submit"
                  className="submit-button db-submit"
                  disabled={loading}
                >
                  {loading ? "ENTRANDO..." : "ENTRAR NA CONTA"}
                  {!loading && <span aria-hidden="true">→</span>}
                </button>
              </form>

              <div className="register-section db-register-section">
                <span>Ainda não tem uma conta?</span>
                <button
                  type="button"
                  className="auth-link db-register-link"
                  onClick={handleGoToRegister}
                >
                  CRIAR CONTA ↗
                </button>
              </div>
            </div>
          )}

          <footer className="form-footer db-form-footer">
            LINK LONGO → LINK CURTO
          </footer>
        </section>
      </div>
    </main>
  );
}

export default App;
