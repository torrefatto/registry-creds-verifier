# OCI Registry Credentials verifier

Does what is written on the can: given a [registry url](#registry-url) and a
user/pass pair, it does two things:

 - verifies if the credentials are valid
 - prints the obtained jwt token (if you want to manually inspect the claims returned)

**NOTE**: It does not verify the jwt token signature.

### Get it

Install:

```
go install github.com/torrefatto/registry-creds-verifier
```

Usage:

```
registry-creds-verifier REGISTRY_URL USERNAME PASSWORD
```

All the positional arguments are mandatory

### registry url

The url to be provided is the one where the [OCI registry API](api) is exposed.

This CLI queries this endpoint (on its `/v2` path) and uses the returned
`WWW-Authenticate` header in the response to know where is the endpoint to
verify the credentials.

For popular registries

 |registry                 |index url              |`WWW-Authenticate` response                                                                |
 |-------------------------|-----------------------|-------------------------------------------------------------------------------------------|
 |docker.io                |https://index.docker.io|`Bearer realm="https://auth.docker.io/token",service="registry.docker.io"`                 |
 |Github Container Registry|https://ghcr.io        |`Bearer realm="https://ghcr.io/token",service="ghcr.io",scope="repository:user/image:pull"`|
 |GCP Container Registry   |https://gcr.io         |`Bearer realm="https://gcr.io/v2/token",service="gcr.io"`                                  |

[api]: https://distribution.github.io/distribution/
