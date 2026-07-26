package metadata

// CapabilityMarker maps a dependency-name pattern to a capability category.
// Matching is case-insensitive substring over the dependency name, so a single
// curated pattern covers variants across ecosystems (e.g. "prisma", "pg",
// "github.com/lib/pq"). This is a deterministic heuristic indicator only.
type CapabilityMarker struct {
	Pattern  string
	Category string
}

var defaultCapabilityMarkers = []CapabilityMarker{
	// database (drivers / clients)
	{Pattern: "postgres", Category: "database"},
	{Pattern: "pg", Category: "database"},
	{Pattern: "mysql", Category: "database"},
	{Pattern: "sqlite", Category: "database"},
	{Pattern: "mongo", Category: "database"},
	{Pattern: "mongoose", Category: "database"},
	{Pattern: "redis", Category: "database"},
	{Pattern: "elasticsearch", Category: "database"},
	{Pattern: "dynamodb", Category: "database"},
	{Pattern: "mariadb", Category: "database"},
	{Pattern: "influxdb", Category: "database"},
	{Pattern: "cassandra", Category: "database"},
	{Pattern: "psycopg", Category: "database"},
	{Pattern: "sqlx", Category: "database"},
	{Pattern: "lib/pq", Category: "database"},

	// orm
	{Pattern: "prisma", Category: "orm"},
	{Pattern: "typeorm", Category: "orm"},
	{Pattern: "sequelize", Category: "orm"},
	{Pattern: "knex", Category: "orm"},
	{Pattern: "gorm", Category: "orm"},
	{Pattern: "drizzle", Category: "orm"},
	{Pattern: "sqlalchemy", Category: "orm"},
	{Pattern: "ent", Category: "orm"},
	{Pattern: "objection", Category: "orm"},
	{Pattern: "mikro-orm", Category: "orm"},

	// auth
	{Pattern: "passport", Category: "auth"},
	{Pattern: "auth0", Category: "auth"},
	{Pattern: "clerk", Category: "auth"},
	{Pattern: "next-auth", Category: "auth"},
	{Pattern: "jsonwebtoken", Category: "auth"},
	{Pattern: "jwt", Category: "auth"},
	{Pattern: "lucia", Category: "auth"},
	{Pattern: "supertokens", Category: "auth"},
	{Pattern: "casbin", Category: "auth"},
	{Pattern: "bcrypt", Category: "auth"},
	{Pattern: "argon2", Category: "auth"},
	{Pattern: "oauth", Category: "auth"},

	// payments
	{Pattern: "stripe", Category: "payments"},
	{Pattern: "paypal", Category: "payments"},
	{Pattern: "razorpay", Category: "payments"},
	{Pattern: "square", Category: "payments"},
	{Pattern: "mollie", Category: "payments"},
	{Pattern: "paddle", Category: "payments"},
	{Pattern: "adyen", Category: "payments"},

	// queue / messaging
	{Pattern: "bullmq", Category: "queue"},
	{Pattern: "bull", Category: "queue"},
	{Pattern: "kafka", Category: "queue"},
	{Pattern: "rabbitmq", Category: "queue"},
	{Pattern: "amqp", Category: "queue"},
	{Pattern: "nsq", Category: "queue"},
	{Pattern: "nats", Category: "queue"},
	{Pattern: "sidekiq", Category: "queue"},
	{Pattern: "celery", Category: "queue"},
	{Pattern: "asynq", Category: "queue"},

	// search
	{Pattern: "meilisearch", Category: "search"},
	{Pattern: "typesense", Category: "search"},
	{Pattern: "algolia", Category: "search"},
	{Pattern: "lunr", Category: "search"},
	{Pattern: "minisearch", Category: "search"},

	// container
	{Pattern: "docker", Category: "container"},
	{Pattern: "dockerode", Category: "container"},
	{Pattern: "testcontainers", Category: "container"},

	// orchestration
	{Pattern: "kubernetes", Category: "orchestration"},
	{Pattern: "k8s.io", Category: "orchestration"},
	{Pattern: "helm", Category: "orchestration"},
	{Pattern: "client-go", Category: "orchestration"},

	// caching
	{Pattern: "memcached", Category: "caching"},
	{Pattern: "lru-cache", Category: "caching"},
	{Pattern: "node-cache", Category: "caching"},

	// monitoring / observability
	{Pattern: "opentelemetry", Category: "monitoring"},
	{Pattern: "prometheus", Category: "monitoring"},
	{Pattern: "datadog", Category: "monitoring"},
	{Pattern: "sentry", Category: "monitoring"},
	{Pattern: "winston", Category: "monitoring"},
	{Pattern: "pino", Category: "monitoring"},
	{Pattern: "go.uber.org/zap", Category: "monitoring"},
}

// DefaultCapabilityMarkers returns a defensive copy of the default capability markers.
func DefaultCapabilityMarkers() []CapabilityMarker {
	markers := make([]CapabilityMarker, len(defaultCapabilityMarkers))
	copy(markers, defaultCapabilityMarkers)
	return markers
}
