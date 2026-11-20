# João Burger fixture

`npm run server` resets the active schema and served media directory, then uses this fixture before opening the public API port. Use a dedicated development database.

Run the HTTP fixture on a disposable PostgreSQL schema:

```sh
AUTH_TEST_DATABASE_URL='postgres://…' go test -race -count=1 ./server/internal/integration -run TestJoaoBurgerHTTPFixture
```

The test starts an HTTP server, invokes `SeedJoaoBurger`, verifies the public menu and account profiles, downloads their image URLs, and drops the schema. The fixture uses `joao-burger` as its API slug and `media/joao_burger` as its image directory.

To reuse `SeedJoaoBurger` against a clean development API, use a platform Administrator access token and start the API with `MEDIA_DIR=server/internal/integration/media` (relative to the repo root). The fixture generates a shared random password and returns it with the fixture accounts for callers that need to sign in.
