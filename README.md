# 🔗 Dulio Shortener

Dulio Shortener is a small full-stack application for creating and managing
short URLs. Users create an account, sign in, save links, and share compact URLs
that redirect visitors to the original destination.

The project contains a Go API built with Echo and SQLite, a React and TypeScript
frontend built with Vite, and an Astro Starlight API-reference site generated
from Markdown and OpenAPI.

## ✨ What it does

- Creates accounts using a username, optional display name, and password.
- Authenticates users and keeps their links private to their account.
- Creates, lists, and deletes short links.
- Redirects public short URLs without loading the frontend.
- Limits each account to 50 links and rejects duplicate destinations.

Implementation and design details are documented in
[ARCHITECTURE.md](ARCHITECTURE.md).

## 🧰 Requirements

For the container-based workflow:

- Docker
- Docker Compose

For native development:

- Go 1.27.2 or newer
- A C compiler available on `PATH`, required by `go-sqlite3`
- Node.js 22.12 or newer, plus npm. Astro 7 in `docs` requires Node 22.12 even
  though the Vite frontend can also build on Node 20.19.

## 🏗️ How to build

Build the API and migration container images from the repository root:

```sh
docker compose build api migrate
```

Compile all Go packages natively:

```sh
go build ./...
```

Install the locked frontend dependencies and create a production build:

```sh
cd frontend
npm ci
npm run build
```

The frontend build is written to `frontend/dist`.

Install the locked documentation dependencies, validate its OpenAPI contract,
and create the static documentation build:

```sh
cd docs
npm ci
npm run check
npm run build
```

The documentation build is written to `docs/dist`.

## 🚀 How to run

### 1️⃣ Configure the project

Copy the example environment file:

```sh
cp .env.example .env
```

PowerShell equivalent:

```powershell
Copy-Item .env.example .env
```

Review `.env` before starting. On Linux, `DULIO_UID` and `DULIO_GID` must
match the user that owns the `data` directory.

### 2️⃣ Add the TLS certificate

Create `data/tls` and place the certificate files expected by the API there:

```text
data/
└── tls/
    ├── origin.pem
    └── origin.key
```

The configured container user must be able to read both files and write to the
`data` directory.

### 3️⃣ Apply the database migrations

Run migrations independently before starting the API:

```sh
docker compose run --rm migrate up
```

Useful migration commands include:

```sh
docker compose run --rm migrate status
docker compose run --rm migrate version
docker compose run --rm migrate down
```

`down` rolls back one migration and should be used intentionally. The API never
applies migrations automatically.

### 4️⃣ Start the API

```sh
docker compose up --build -d api
```

View its logs or stop it with:

```sh
docker compose logs -f api
docker compose down
```

The API uses the host port configured by `DULIO_HTTP_PORT`, which defaults to
`443`.

### 5️⃣ Start the frontend

The frontend is not part of the Compose file. Start its development server in a
separate terminal:

```sh
cd frontend
npm ci
npm run dev
```

Vite prints the local development URL after it starts.

## 🗃️ Native migrations

To run the migration tool without Docker in a POSIX shell:

```sh
DATABASE_PATH=./data/dulio.db go run ./database/cmd/migrate up
```

In PowerShell:

```powershell
$env:DATABASE_PATH = ".\data\dulio.db"
go run ./database/cmd/migrate up
```

Replace `up` with `status`, `version`, or `down` when needed.

## 🧪 Verification

Run the backend checks from the repository root:

```sh
go test ./...
go test -race ./...
go vet ./...
```

Run the frontend checks from `frontend`:

```sh
npm run lint
npm run build
```

Run the documentation checks from `docs`:

```sh
npm run check
npm run build
npm audit
```

## ☁️ Deployment

The Go API currently runs on Oracle Cloud Infrastructure (OCI). Its origin,
network, TLS certificate, and database provisioning remain environment-specific
and are not automated by this repository.

### API documentation on Cloudflare Pages

Cloudflare Pages supports Astro directly, including the Legacy Pages build
configuration. Create a Pages project from this Git repository with:

| Setting | Value |
| --- | --- |
| Production branch | `master` |
| Framework preset | `Astro` |
| Root directory | `docs` |
| Build command | `npm run build` |
| Build output directory | `dist` |

If the Legacy Pages interface does not list Astro, select no framework preset
and enter the same root, command, and output directory manually. This is a
fully static Astro build and does not need `@astrojs/cloudflare`, Pages
Functions, Workers, or a Wrangler configuration. Set `NODE_VERSION` to a value
of at least `22.12.0` if the Pages build image does not honor the package
`engines` field automatically.

After the first successful deployment:

1. Open the Pages project and select **Custom domains**.
2. Add `api-reference.3dreamstudio.com.br` and activate it.
3. Because `3dreamstudio.com.br` is already a Cloudflare-managed zone,
   Cloudflare should create the required CNAME record automatically.
4. Verify the custom domain is active and that `/openapi.yaml` and
   `/api-reference/` are reachable.

Attach the custom domain through the Pages project before creating or changing
DNS manually. A CNAME pointed at `*.pages.dev` without the Pages custom-domain
association can fail with HTTP 522.

This repository prepares and verifies the deployment artifact but does not
create the Cloudflare Pages project or mutate external DNS automatically.
