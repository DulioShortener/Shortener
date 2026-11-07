
import { useCallback, useEffect, useState, type FormEvent } from "react";
import type { LoginResponse } from "../services/auth";
import {
  getLinks,
  createLink,
  deleteLink,
  type ShortLink,
} from "../services/links";

interface DashboardProps {
  session: LoginResponse;
  onLogout: () => void;
}

function LinkIcon() {
  return (
    <svg
      width="23"
      height="23"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
      <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
    </svg>
  );
}

function Dashboard({ session, onLogout }: DashboardProps) {
  const name = session.user.display_name || session.user.username;

  const [links, setLinks] = useState<ShortLink[]>([]);
  const [url, setUrl] = useState("");
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const loadLinks = useCallback(async () => {
    try {
      setLoading(true);
      setError("");
      const result = await getLinks(session.token);
      setLinks(result);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Erro ao carregar links."
      );
    } finally {
      setLoading(false);
    }
  }, [session.token]);

  useEffect(() => {
    void loadLinks();
  }, [loadLinks]);

  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setMessage("");

    try {
      const parsed = new URL(url.trim());

      if (!["http:", "https:"].includes(parsed.protocol)) {
        setError("Utilize um endereço HTTP ou HTTPS.");
        return;
      }

      if (new TextEncoder().encode(url.trim()).length > 4096) {
        setError("A URL excede o limite de 4096 bytes.");
        return;
      }
    } catch {
      setError("Digite uma URL válida, incluindo https://.");
      return;
    }

    if (links.length >= 50) {
      setError("Você atingiu o limite de 50 links.");
      return;
    }

    setCreating(true);

    try {
      const newLink = await createLink(url.trim(), session.token);

      setLinks((current) => [newLink, ...current]);
      setUrl("");
      setMessage("Link encurtado com sucesso!");
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Erro ao criar link."
      );
    } finally {
      setCreating(false);
    }
  }

  async function handleDelete(id: string) {
    if (!window.confirm("Deseja realmente excluir este link?")) {
      return;
    }

    setDeletingId(id);
    setError("");
    setMessage("");

    try {
      await deleteLink(id, session.token);
      setLinks((current) => current.filter((link) => link.id !== id));
      setMessage("Link excluído com sucesso.");
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Erro ao excluir link."
      );
    } finally {
      setDeletingId(null);
    }
  }

  async function handleCopy(link: ShortLink) {
    try {
      await navigator.clipboard.writeText(link.short_url);
      setCopiedId(link.id);
      setError("");
      setMessage("Link copiado!");

    } catch {
      setError("Não foi possível copiar o link.");
    }
  }

  return (
    <main className="dashboard">
      <header className="dashboard-header">
        <div className="dashboard-brand">
          <div className="brand-icon">
            <LinkIcon />
          </div>
          <strong>Dulio.</strong>
        </div>

        <div className="dashboard-user">
          <div className="user-info">
            <span className="user-greeting">Conectado como</span>
            <strong>{name}</strong>
          </div>

          <div className="user-avatar">
            {name.charAt(0).toUpperCase()}
          </div>

          <button className="logout-button" onClick={onLogout}>
            Sair
          </button>
        </div>
      </header>

      <section className="dashboard-content">
        <div className="dashboard-welcome">
          <div className="hero-badge">
            <span className="status-dot" />
            SEU PAINEL DE CONTROLE
          </div>

          <h1>
            Bem-vindo, <span>{name}!</span>
          </h1>
          <p>Crie, organize e compartilhe seus links em um só lugar.</p>
        </div>

        <div className="dashboard-stats">
          <div className="stat-card">
            <span className="stat-label">LINKS CADASTRADOS</span>
            <strong>{loading ? "..." : links.length}</strong>
            <p>De até 50 links por conta</p>
          </div>

          <div className="stat-card">
            <span className="stat-label">SUA CONTA</span>
            <strong className="stat-active">Ativa</strong>
            <p>Sessão autenticada</p>
          </div>
        </div>

        {error && (
          <div className="error-message" role="alert">
            {error}
          </div>
        )}

        {message && (
          <div className="success-message" role="status">
            ✓ {message}
          </div>
        )}

        <div className="dashboard-panel">
          <div className="panel-heading">
            <div className="panel-icon">
              <LinkIcon />
            </div>
            <div>
              <span className="welcome-label">ENCURTADOR DE LINKS</span>
              <h2>Crie um novo link</h2>
            </div>
          </div>

          <p>
            Cole sua URL original para gerar um link curto
            e fácil de compartilhar.
          </p>

          <form className="dashboard-url-form" onSubmit={handleCreate}>
            <input
              type="url"
              placeholder="https://exemplo.com/sua-url"
              value={url}
              onChange={(event) => setUrl(event.target.value)}
              required
              disabled={creating || loading || links.length >= 50}
            />

            <button
              type="submit"
              disabled={creating || loading || links.length >= 50}
            >
              {creating ? "Encurtando..." : "Encurtar link →"}
            </button>
          </form>

          <span className="dashboard-hint">
            Você pode cadastrar até 50 links na sua conta.
          </span>
        </div>

        <div className="dashboard-panel">
          <div className="panel-heading">
            <div className="panel-icon">
              <LinkIcon />
            </div>
            <div>
              <span className="welcome-label">GERENCIAMENTO</span>
              <h2>Meus links</h2>
            </div>
          </div>

          {loading ? (
            <p>Carregando seus links...</p>
          ) : links.length === 0 ? (
            <div className="empty-links">
              <div className="empty-icon">
                <LinkIcon />
              </div>
              <h3>Nenhum link por aqui</h3>
              <p>Crie seu primeiro link para começar.</p>
              {error && (
                <button className="logout-button" onClick={() => void loadLinks()}>
                  Tentar novamente
                </button>
              )}
            </div>
          ) : (
            <div className="links-list">
              {links.map((link) => (
                <article className="link-item" key={link.id}>
                  <div className="link-details">
                    <a
                      className="short-link-text"
                      href={link.short_url}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      <LinkIcon />
                      {link.short_url}
                    </a>

                    <p title={link.target_url}>
                      {link.target_url}
                    </p>

                    <span className="link-date">
                      Criado em{" "}
                      {new Date(link.created_at).toLocaleDateString("pt-BR")}
                    </span>
                  </div>

                  <div className="link-actions">
                    <button
                      type="button"
                      className="copy-link-button"
                      onClick={() => void handleCopy(link)}
                    >
                      {copiedId === link.id ? "✓ Copiado" : "Copiar"}
                    </button>

                    <button
                      type="button"
                      className="delete-link-button"
                      disabled={deletingId === link.id}
                      onClick={() => void handleDelete(link.id)}
                    >
                      {deletingId === link.id ? "Excluindo..." : "Excluir"}
                    </button>
                  </div>
                </article>
              ))}
            </div>
          )}
        </div>
      </section>

      <footer className="dashboard-footer">
        © {new Date().getFullYear()} Dulio Shortener
      </footer>
    </main>
  );
}

export default Dashboard;
