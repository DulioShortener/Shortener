import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  type FormEvent,
} from "react";

import type { LoginResponse } from "../services/auth";

import {
  getReviews,
  getMyReview,
  createReview,
  getReplies,
  createReply,
  deleteReview,
  type Review,
} from "../services/reviews";

interface ReviewsProps {
  session: LoginResponse;
  onBack: () => void;
}

function getErrorMessage(error: unknown): string {
  return error instanceof Error
    ? error.message
    : "Ocorreu um erro inesperado.";
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString("pt-BR");
}

function DeleteModal({
  open,
  loading,
  onCancel,
  onConfirm,
}: {
  open: boolean;
  loading: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  if (!open) return null;

  return (
    <div className="custom-modal-overlay" role="dialog" aria-modal="true">
      <div className="custom-modal">
        <div className="custom-modal-icon">!</div>

        <h3>Excluir conteúdo</h3>

        <p>
          Tem certeza que deseja excluir este conteúdo?
          Essa ação não poderá ser desfeita no frontend.
        </p>

        <div className="custom-modal-actions">
          <button
            type="button"
            className="review-secondary-button"
            onClick={onCancel}
            disabled={loading}
          >
            Cancelar
          </button>

          <button
            type="button"
            className="review-danger-button"
            onClick={onConfirm}
            disabled={loading}
          >
            {loading ? "Excluindo..." : "Excluir"}
          </button>
        </div>
      </div>
    </div>
  );
}

function Comment({
  item,
  token,
  userId,
  onChanged,
  depth = 0,
}: {
  item: Review;
  token: string;
  userId: string;
  onChanged: () => Promise<void>;
  depth?: number;
}) {
  const [expanded, setExpanded] = useState(false);
  const [replying, setReplying] = useState(false);
  const [body, setBody] = useState("");
  const [replies, setReplies] = useState<Review[]>([]);
  const [cursor, setCursor] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);
  const [loading, setLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState("");
  const [showDeleteModal, setShowDeleteModal] = useState(false);

  const isMine = item.author_id === userId;

  const visibleReplies = useMemo(
    () => replies.filter((reply) => reply.body !== null),
    [replies]
  );

  const hasChildren = item.has_replies || visibleReplies.length > 0;

  const loadReplies = useCallback(
    async (nextCursor?: string) => {
      setLoading(true);
      setError("");

      try {
        const page = await getReplies(token, item.id, nextCursor);
        const filteredItems = page.items.filter((reply) => reply.body !== null);

        if (nextCursor) {
          setReplies((current) => [...current, ...filteredItems]);
        } else {
          setReplies(filteredItems);
        }

        setHasMore(page.has_more);
        setCursor(page.next_cursor);
      } catch (err) {
        setError(getErrorMessage(err));
      } finally {
        setLoading(false);
      }
    },
    [token, item.id]
  );

  async function toggleReplies() {
    if (expanded) {
      setExpanded(false);
      return;
    }

    setExpanded(true);
    await loadReplies();
  }

  async function sendReply(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!body.trim() || Array.from(body).length > 2000) {
      setError("Escreva uma resposta de 1 a 2000 caracteres.");
      return;
    }

    setSending(true);
    setError("");

    try {
      await createReply(token, item.id, body.trim());
      setBody("");
      setReplying(false);
      setExpanded(true);
      await loadReplies();
      await onChanged();
    } catch (err) {
      setError(getErrorMessage(err));
    } finally {
      setSending(false);
    }
  }

  async function confirmDelete() {
    setDeleting(true);
    setError("");

    try {
      await deleteReview(token, item.id);
      setShowDeleteModal(false);
      await onChanged();
    } catch (err) {
      setError(getErrorMessage(err));
    } finally {
      setDeleting(false);
    }
  }

  if (item.body === null) {
    return null;
  }

  return (
    <>
      <article
        className={`review-comment ${depth > 0 ? "review-reply" : ""}`}
      >
        <div className="review-comment-header">
          <div className="review-author">
            <div className="review-avatar">
              {isMine ? "V" : "U"}
            </div>

            <div>
              <strong>
                {isMine ? "Você" : "Membro da comunidade"}
              </strong>
              <span>{formatDate(item.created_at)}</span>
            </div>
          </div>

          {typeof item.rating === "number" && (
            <div
              className="review-stars-display"
              aria-label={`${item.rating} de 5 estrelas`}
            >
              {"★".repeat(item.rating)}
              <span>{"★".repeat(5 - item.rating)}</span>
            </div>
          )}
        </div>

        <p className="review-body">{item.body}</p>

        <div className="review-actions">
          <button
            type="button"
            onClick={() => setReplying(!replying)}
          >
            Responder
          </button>

          {hasChildren && (
            <button
              type="button"
              onClick={() => void toggleReplies()}
            >
              {expanded ? "Ocultar respostas" : "Ver respostas"}
            </button>
          )}

          {isMine && (
            <button
              type="button"
              className="review-delete"
              onClick={() => setShowDeleteModal(true)}
            >
              Excluir
            </button>
          )}
        </div>

        {replying && (
          <form className="reply-form" onSubmit={sendReply}>
            <textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder="Escreva sua resposta..."
              maxLength={2000}
              required
              rows={3}
            />

            <div className="reply-form-actions">
              <button
                type="button"
                className="review-secondary-button"
                onClick={() => setReplying(false)}
              >
                Cancelar
              </button>

              <button
                type="submit"
                className="review-primary-button"
                disabled={sending}
              >
                {sending ? "Enviando..." : "Enviar resposta"}
              </button>
            </div>
          </form>
        )}

        {error && (
          <div className="error-message" role="alert">
            {error}
          </div>
        )}

        {expanded && (
          <div className="review-replies">
            {visibleReplies.map((reply) => (
              <Comment
                key={reply.id}
                item={reply}
                token={token}
                userId={userId}
                onChanged={async () => {
                  await loadReplies();
                  await onChanged();
                }}
                depth={depth + 1}
              />
            ))}

            {loading && <p>Carregando respostas...</p>}

            {!loading && hasMore && cursor && (
              <button
                className="review-secondary-button"
                onClick={() => void loadReplies(cursor)}
              >
                Carregar mais respostas
              </button>
            )}
          </div>
        )}
      </article>

      <DeleteModal
        open={showDeleteModal}
        loading={deleting}
        onCancel={() => setShowDeleteModal(false)}
        onConfirm={() => void confirmDelete()}
      />
    </>
  );
}

function Reviews({ session, onBack }: ReviewsProps) {
  const [rating, setRating] = useState(5);
  const [body, setBody] = useState("");
  const [reviews, setReviews] = useState<Review[]>([]);
  const [myReview, setMyReview] = useState<Review | null>(null);

  const [cursor, setCursor] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);

  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [sending, setSending] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const refresh = useCallback(async () => {
    const [page, mine] = await Promise.all([
      getReviews(session.token),
      getMyReview(session.token),
    ]);

    const visibleReviews = page.items.filter((review) => review.body !== null);

    setReviews(visibleReviews);
    setMyReview(mine && mine.body !== null ? mine : null);
    setCursor(page.next_cursor);
    setHasMore(page.has_more);
  }, [session.token]);

  useEffect(() => {
    let active = true;

    Promise.all([
      getReviews(session.token),
      getMyReview(session.token),
    ])
      .then(([page, mine]) => {
        if (!active) return;

        const visibleReviews = page.items.filter(
          (review) => review.body !== null
        );

        setReviews(visibleReviews);
        setMyReview(mine && mine.body !== null ? mine : null);
        setCursor(page.next_cursor);
        setHasMore(page.has_more);
      })
      .catch((err) => {
        if (active) setError(getErrorMessage(err));
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [session.token]);

  async function handlePublish(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!body.trim() || Array.from(body).length > 2000) {
      setError("A avaliação deve ter de 1 a 2000 caracteres.");
      return;
    }

    setSending(true);
    setError("");
    setSuccess("");

    try {
      await createReview(session.token, body.trim(), rating);
      setBody("");
      setSuccess("Sua avaliação foi publicada!");
      await refresh();
    } catch (err) {
      setError(getErrorMessage(err));
      try {
        await refresh();
      } catch {
        // mantém o erro original
      }
    } finally {
      setSending(false);
    }
  }

  async function loadMore() {
    if (!cursor || loadingMore) return;

    setLoadingMore(true);
    setError("");

    try {
      const page = await getReviews(session.token, cursor);
      const visibleReviews = page.items.filter((review) => review.body !== null);

      setReviews((current) => [...current, ...visibleReviews]);
      setCursor(page.next_cursor);
      setHasMore(page.has_more);
    } catch (err) {
      setError(getErrorMessage(err));
    } finally {
      setLoadingMore(false);
    }
  }

  async function refreshAfterDelete() {
    setSuccess("");
    await refresh();
  }

  return (
    <main className="dashboard reviews-page">
      <section className="dashboard-content">
        <div className="dashboard-welcome">
          <div className="hero-badge">
            ★ AVALIAÇÕES DA COMUNIDADE
          </div>

          <h1>
            Sua opinião <span>importa.</span>
          </h1>

          <p>
            Conte como está sendo sua experiência com o Dulio
            Shortener e converse com outros usuários.
          </p>
        </div>

        <div className="reviews-top-actions">
          <button
            className="logout-button"
            type="button"
            onClick={onBack}
          >
            ← Voltar ao Dashboard
          </button>
        </div>

        {error && (
          <div className="error-message" role="alert">
            {error}
          </div>
        )}

        {success && (
          <div className="success-message" role="status">
            ✓ {success}
          </div>
        )}

        {!loading && !myReview && (
          <div className="dashboard-panel review-form-card">
            <span className="welcome-label">
              COMPARTILHE SUA EXPERIÊNCIA
            </span>

            <h2>Avalie o Dulio Shortener</h2>

            <form onSubmit={handlePublish}>
              <label>Quantas estrelas você daria?</label>

              <div
                className="review-rating-input"
                aria-label="Escolha uma nota de 1 a 5"
              >
                {[1, 2, 3, 4, 5].map((value) => (
                  <button
                    type="button"
                    key={value}
                    className={value <= rating ? "selected" : ""}
                    onClick={() => setRating(value)}
                    aria-label={`${value} estrelas`}
                    aria-pressed={rating === value}
                  >
                    ★
                  </button>
                ))}
              </div>

              <label htmlFor="review-body">
                Sua avaliação
              </label>

              <textarea
                id="review-body"
                value={body}
                onChange={(e) => setBody(e.target.value)}
                placeholder="Conte o que achou do encurtador..."
                maxLength={2000}
                required
                rows={5}
              />

              <div className="review-form-footer">
                <span>{Array.from(body).length}/2000 caracteres</span>

                <button
                  type="submit"
                  className="review-primary-button"
                  disabled={sending}
                >
                  {sending ? "Publicando..." : "Publicar avaliação →"}
                </button>
              </div>
            </form>
          </div>
        )}

        {myReview && (
          <div className="dashboard-panel">
            <span className="welcome-label">
              SUA AVALIAÇÃO
            </span>
            <p>
              Você já publicou sua avaliação.
              É permitido manter uma avaliação ativa por conta.
            </p>
          </div>
        )}

        <div className="dashboard-panel">
          <div className="panel-heading">
            <div className="panel-icon">★</div>
            <div>
              <span className="welcome-label">
                COMUNIDADE
              </span>
              <h2>Avaliações dos usuários</h2>
            </div>
          </div>

          {loading ? (
            <p>Carregando avaliações...</p>
          ) : reviews.length === 0 ? (
            <div className="empty-links">
              <h3>Nenhuma avaliação encontrada</h3>
              <p>Seja o primeiro a compartilhar sua opinião!</p>
            </div>
          ) : (
            <div className="reviews-list">
              {reviews.map((review) => (
                <Comment
                  key={review.id}
                  item={review}
                  token={session.token}
                  userId={session.user.id}
                  onChanged={refreshAfterDelete}
                />
              ))}
            </div>
          )}

          {!loading && hasMore && cursor && (
            <button
              className="review-more-button"
              disabled={loadingMore}
              onClick={() => void loadMore()}
            >
              {loadingMore
                ? "Carregando..."
                : "Carregar mais avaliações"}
            </button>
          )}
        </div>
      </section>

      <footer className="dashboard-footer">
        © {new Date().getFullYear()} Dulio Shortener
      </footer>
    </main>
  );
}

export default Reviews;