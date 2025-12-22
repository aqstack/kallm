# kallm

**Kubernetes-native LLM Semantic Cache**

kallm is a drop-in proxy that caches LLM API responses using semantic similarity, reducing costs and latency for repeated or similar queries.

## Features

- **Semantic Caching** - Cache hits for semantically similar prompts, not just exact matches
- **OpenAI-Compatible** - Drop-in replacement proxy for OpenAI API
- **Configurable Threshold** - Tune similarity sensitivity (0.0-1.0)
- **TTL Support** - Time-based cache expiration
- **Zero Dependencies** - Single binary, no external database required
- **Kubernetes-Ready** - Designed for cloud-native deployments

## How It Works

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│    kallm    │────▶│  LLM API    │
│  (app/pod)  │◀────│   (proxy)   │◀────│ (OpenAI/..) │
└─────────────┘     └──────┬──────┘     └─────────────┘
                           │
                    ┌──────▼──────┐
                    │ Vector Store │
                    │  (embeddings)│
                    └─────────────┘
```

1. Incoming request is converted to an embedding
2. Cache is searched for semantically similar previous requests
3. If similarity exceeds threshold → return cached response
4. Otherwise → forward to upstream, cache response

## Quick Start

### Using Docker

```bash
# Set your OpenAI API key
export OPENAI_API_KEY=sk-...

# Run kallm
docker run -p 8080:8080 -e OPENAI_API_KEY=$OPENAI_API_KEY ghcr.io/aqstack/kallm:latest
```

### Using Docker Compose

```bash
# Clone the repo
git clone https://github.com/aqstack/kallm.git
cd kallm

# Set your API key
export OPENAI_API_KEY=sk-...

# Run
docker-compose up
```

### Building from Source

```bash
# Clone
git clone https://github.com/aqstack/kallm.git
cd kallm

# Build
make build

# Run
export OPENAI_API_KEY=sk-...
./bin/kallm
```

## Usage

Point your OpenAI client to kallm instead of the OpenAI API:

```python
from openai import OpenAI

# Point to kallm proxy
client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="your-api-key"  # or use OPENAI_API_KEY env var
)

response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "What is the capital of France?"}]
)

# Check cache status in response headers
# X-Kallm-Cache: HIT or MISS
# X-Kallm-Similarity: 0.9823 (if HIT)
```

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `OPENAI_API_KEY` | (required) | OpenAI API key |
| `OPENAI_BASE_URL` | `https://api.openai.com/v1` | Upstream API URL |
| `KALLM_PORT` | `8080` | Server port |
| `KALLM_HOST` | `0.0.0.0` | Server host |
| `KALLM_SIMILARITY_THRESHOLD` | `0.95` | Minimum similarity for cache hit (0.0-1.0) |
| `KALLM_CACHE_TTL` | `24h` | Cache entry time-to-live |
| `KALLM_MAX_CACHE_SIZE` | `10000` | Maximum cache entries |
| `KALLM_EMBEDDING_MODEL` | `text-embedding-3-small` | Embedding model |
| `KALLM_LOG_JSON` | `false` | JSON log format |

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `POST /v1/chat/completions` | Chat completions (cached) |
| `GET /health` | Health check |
| `GET /stats` | Cache statistics |
| `* /v1/*` | Other OpenAI endpoints (passthrough) |

## Cache Statistics

```bash
curl http://localhost:8080/stats
```

```json
{
  "total_entries": 150,
  "total_hits": 1234,
  "total_misses": 567,
  "hit_rate": 0.685,
  "estimated_saved_usd": 1.234
}
```

## Tuning the Similarity Threshold

The `KALLM_SIMILARITY_THRESHOLD` controls how similar a query must be to trigger a cache hit:

| Threshold | Behavior |
|-----------|----------|
| `0.99` | Nearly exact matches only |
| `0.95` | Very similar queries (recommended) |
| `0.90` | Moderate similarity |
| `0.85` | Loose matching (may return less relevant) |

## Roadmap

- [ ] Redis/Qdrant backend for persistence
- [ ] Kubernetes Helm chart
- [ ] Prometheus metrics
- [ ] CRD for cache configuration
- [ ] Namespace-based isolation
- [ ] Cache warming
- [ ] Support for Anthropic, Gemini APIs

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

MIT License - see [LICENSE](LICENSE) for details.
