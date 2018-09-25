package config

var (
	SERVICE_WHITELIST = []string{
		"mysql", "p-mysql", "p.mysql",
		"rabbit", "rabbitmq", "p-rabbitmq", "p.rabbitmq",
		"redis", "p-redis", "p.redis",
		"p-circuit-breaker-dashboard", "p-config-server", "p-service-registry",
	}
)
