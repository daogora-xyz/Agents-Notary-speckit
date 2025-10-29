  Milestone 2: x402 MCP Server
  /speckit.specify x402 payment MCP server: Build MCP server using mcp-go with 5 tools - create_payment_requirement
  (generates x402 JSON), verify_payment (EIP-3009 signature validation), settle_payment (calls facilitator API),
  generate_browser_link (MetaMask deep links), encode_payment_for_qr (EIP-681 format). Support base, base-sepolia,
  arbitrum networks. Reference: docs/OVERVIEW.md Milestone 2, Section 2.3.2 tool schemas, Tasks T005-T011.

  Milestone 3: Circular Protocol MCP Server (Week 2-3)

  /speckit.specify Circular Protocol MCP server: Build MCP server using mcp-go with 4 tools for blockchain
  certification operations. Tools: get_wallet_nonce (fetches current nonce via Circular_GetWalletNonce_ API),
  certify_data (constructs C_TYPE_CERTIFICATE transaction with Secp256k1 signing, posts via Circular_AddTransaction_),
  get_transaction_status (polls transaction status until "Executed"), get_certification_proof (extracts block ID,
  timestamp, generates explorer URL). Must handle Circular Protocol HTTP REST API, implement transaction ID calculation
   (sha256 of From+To+Payload+Timestamp), support testnet and mainnet. Testing: end-to-end certification on testnet,
  measure confirmation time. Reference: docs/OVERVIEW.md Milestone 3 (Week 2-3), Section 2.3.2 tool schemas, Tasks
  T012-T017.

  Milestone 4: Data Quote & QR Code MCP Servers (Week 3)

  /speckit.specify Data quote and QR code MCP servers: Build two MCP servers using mcp-go. 1) data-quote-mcp-server
  with 3 tools: check_data_size (decodes base64/hex, calculates bytes, formats as KB/MB), get_cirx_price (calls
  CoinGecko API /simple/price?ids=circular, caches for 5 min), calculate_quote (formula: (4 CIRX × price) × (1 +
  margin%), returns detailed breakdown). 2) qr-code-mcp-server with 3 tools: generate_qr_ascii (creates
  terminal-friendly QR using █ character), generate_qr_image (generates PNG/SVG with base64 encoding),
  encode_x402_to_qr (formats payment data as EIP-681 URI with callback URL). Use github.com/skip2/go-qrcode library.
  Testing: unit tests for all tools, integration test with mock CoinGecko API. Reference: docs/OVERVIEW.md Milestone 4
  (Week 3), Tasks T018-T025.

  Milestone 5: MCP Host Proxy Core (Week 4)

  /speckit.specify certify.ar4s.com MCP Host proxy core infrastructure: Initialize HTTP server using Gin or Echo with
  health check endpoint /health. Implement MCP client layer connecting to all 4 MCP servers (x402, circular-protocol,
  data-quote, qr-code) via stdio transport with connection pooling and automatic reconnection. Build database layer
  with PostgreSQL connection pool and CRUD operations for certification_requests, payments, certifications tables.
  Implement Redis layer with cache functions for CIRX price (5 min TTL) and payment status. Create middleware:
  authMiddleware (API key validation), rateLimitMiddleware (Redis-based, 10 req/min), corsMiddleware, loggingMiddleware
   (structured logs with Zap). Configuration via config.yaml + environment variables. Testing: integration tests for
  DB/Redis, middleware unit tests, verify all 4 MCP server connections. Reference: docs/OVERVIEW.md Milestone 5 (Week
  4), Section 2.5 error handling, Tasks T026-T030.

  Milestone 6: MCP Host Proxy API Handlers (Week 4-5)

  /speckit.specify HTTP API endpoints for certification workflow: Implement 4 REST endpoints. 1) POST /v1/quote: parse
  request, call data-quote-mcp tools (check_data_size, get_cirx_price, calculate_quote), return quote response. 2) POST
   /v1/certify (initial): validate request, check request_id idempotency, get quote, call
  x402-mcp.create_payment_requirement, handle client_type (browser→generate_browser_link, mobile→qr-code-mcp tools),
  return 402 Payment Required, store in database. 3) POST /v1/certify (with payment): parse X-PAYMENT header, call
  x402-mcp.verify_payment, if valid call settle_payment, store payment record, trigger async certification workflow,
  return 202 Accepted or 200 OK. 4) GET /v1/status/:id: fetch request/payment/certification from database, build status
   response. 5) GET /v1/qr/:id: fetch request, call qr-code-mcp tools, return QR code as PNG/SVG/ASCII. Testing:
  integration tests for each endpoint covering success and error paths. Reference: docs/OVERVIEW.md Milestone 6 (Week
  4-5), Section 2.4 API specs, Tasks T031-T035.

  Milestone 7: Orchestration & State Machine (Week 5)

  /speckit.specify Certification workflow orchestration with state machine: Implement state machine with states
  (initiated, quoted, payment_pending, payment_verified, certifying, completed, failed) and transition validation,
  store state in database, emit state change events for monitoring. Build certification workflow: after payment
  settled, call circular-protocol-mcp.get_wallet_nonce, sign transaction locally with service wallet, call
  certify_data, poll get_transaction_status until "Executed" or timeout (30s), call get_certification_proof, update
  database, trigger webhook callback if provided. Implement retry queue as background goroutine: scan for failed
  certifications, exponential backoff (5s, 10s, 20s, 40s, max 60s), max 10 attempts, move to dead letter queue after
  exhaustion. Implement webhook callbacks: HTTP POST to callback_url with certification proof, HMAC signature, retry on
   failure (3 attempts). Testing: integration tests for full workflow, retry queue behavior, webhook delivery.
  Reference: docs/OVERVIEW.md Milestone 7 (Week 5), Section 2.5 error handling, Tasks T036-T039.

  Milestone 8: Monitoring & Operations (Week 6)

  /speckit.specify Operational monitoring and observability infrastructure: Implement Prometheus metrics exposed at
  /metrics endpoint covering: certification success rate, payment verification rate, average certification time, CIRX
  wallet balance, HTTP request latency, error rates by endpoint. Build CIRX wallet balance monitor as background job
  running every 5 minutes, update cirx_wallet_balance metric, trigger alert if balance < 100 CIRX. Integrate Zap
  structured JSON logger with levels (DEBUG, INFO, WARN, ERROR), ensure no sensitive data logged (private keys, payment
   headers, PII), log all certification request state transitions with correlation IDs. Create Grafana dashboard JSON
  with panels: certification success rate (7-day trend), average certification time (P50, P95, P99), payment
  verification rate, CIRX balance, HTTP latency, error rate by endpoint. Write Prometheus Alertmanager rules for: low
  CIRX balance (<100), high certification failure rate (>5%), high payment verification failure rate (>10%), slow
  certification time (P95 >15s). Configure PagerDuty or Slack integration. Testing: verify metrics collection, test
  alert firing. Reference: docs/OVERVIEW.md Milestone 8 (Week 6), Section 2.7 monitoring specs, Tasks T040-T044.

  Milestone 9: Testing & QA (Week 7)

  /speckit.specify Comprehensive testing and quality assurance: Achieve 80%+ unit test coverage across all packages,
  95%+ for critical paths (payment, certification), 100% for security code (signature verification, key management).
  Run go test ./... -cover and fix failing tests. Build integration tests for full flow with all 4 MCP servers using
  testnet blockchains, verify end-to-end quote→payment→certification workflow, test all error paths (payment
  verification failures, settlement timeouts, certification retries). Implement load testing using k6 or vegeta: test
  100 concurrent requests, measure P95 latency (target <10s), identify bottlenecks, validate throughput meets 99.5%
  uptime SLA. Conduct security testing: run gosec for vulnerability scanning, test API key authentication bypass
  attempts, test payment signature forgery attempts, verify rate limiting effectiveness (10 req/min), test SQL
  injection resistance, validate no sensitive data in logs. Success criteria: all tests pass, coverage targets met, P95
   latency <10s, no critical vulnerabilities. Reference: docs/OVERVIEW.md Milestone 9 (Week 7), Section 3.3 success
  criteria, Tasks T045-T048.

  Milestone 10: Deployment (Week 8)

  /speckit.specify Production deployment infrastructure and CI/CD: Create Docker images with multi-stage builds for
  size optimization: Dockerfile.proxy (certify.ar4s.com HTTP server), Dockerfile.mcp-server (generic template for 4 MCP
   servers), test local builds. Write Kubernetes manifests: Deployment for proxy (3 replicas with HPA), Deployments for
   4 MCP servers, Service manifests for proxy, ConfigMap for config.yaml, Secret for wallet keys (encrypted),
  HorizontalPodAutoscaler for auto-scaling, test on local minikube. Build GitHub Actions CI/CD pipelines:
  .github/workflows/ci.yaml (run tests on PR, lint with golangci-lint, build Docker images, verify coverage),
  .github/workflows/deploy.yaml (deploy to staging on merge to main, deploy to production on tag release, run smoke
  tests). Deploy to testnet staging environment using base-sepolia and Circular Protocol testnet, run smoke tests,
  verify monitoring dashboards. Deploy to mainnet production: purchase CIRX for production wallet, update config for
  mainnet networks (base, arbitrum, circular-mainnet), deploy, run smoke tests, monitor for 24 hours. Finalize
  documentation: API.md (OpenAPI spec), DEPLOYMENT.md (deployment procedures), RUNBOOK.md (operations procedures),
  update README.md. Success: production running, all 3 user workflows functional (agent, browser, mobile), monitoring
  active. Reference: docs/OVERVIEW.md Milestone 10 (Week 8), Tasks T049-T054.
