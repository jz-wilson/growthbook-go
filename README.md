# growthbook-go

A small, typed Go client for the [GrowthBook](https://www.growthbook.io/)
REST API. It covers the endpoints needed to manage projects, environments,
features (including targeting rules), and SDK connections, and is shared by
[provider-growthbook](https://github.com/jz-wilson/provider-growthbook)
(Crossplane) and
[terraform-provider-growthbook](https://github.com/jz-wilson/terraform-provider-growthbook).

```go
import growthbook "github.com/jz-wilson/growthbook-go"

c, err := growthbook.New(growthbook.Credentials{
    APIKey: os.Getenv("GROWTHBOOK_API_KEY"),
    APIURL: "https://growthbook.example.com/api", // omit for GrowthBook Cloud
})
p, err := c.CreateProject(ctx, growthbook.ProjectRequest{Name: "web"})
```

Every optional request field is a pointer or `omitempty` slice so that
"unset" is distinguishable from "empty". Errors from the API are
`*APIError` with the HTTP status and GrowthBook's message; `IsNotFound`
also recognises GrowthBook's 400 "Could not find ..." responses, and
`IsArchiveRequired` recognises the 403 returned when deleting a live
feature.

## Testing

Unit tests run against `httptest` servers: `go test ./...`.

Integration tests run against a real GrowthBook and are behind the
`integration` build tag:

```bash
docker compose -f e2e/docker-compose.yml up -d --wait
export GROWTHBOOK_API_KEY=$(e2e/bootstrap.sh)
export GROWTHBOOK_API_URL=http://localhost:3100/api
go test -tags integration ./...
```

An unlicensed GrowthBook enforces the free plan (one project, no custom
environments) with HTTP 402; the tests fall back to update-and-revert on
existing objects where creation is refused.

## License

Apache-2.0.
