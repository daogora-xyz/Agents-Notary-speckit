{ pkgs, ... }:

{
  project.name = "certify-platform";

  # Load environment variables from .env if it exists
  services = {
    postgres = {
      service.image = "postgres:16-alpine";
      service.container_name = "certify-postgres";
      service.environment = {
        POSTGRES_USER = "\${POSTGRES_USER:-certify}";
        POSTGRES_PASSWORD = "\${POSTGRES_PASSWORD:-dev_password_change_in_production}";
        POSTGRES_DB = "\${POSTGRES_DB:-certify_platform}";
      };
      service.ports = [
        "\${POSTGRES_PORT:-5432}:5432"
      ];
      service.volumes = [
        "postgres_data:/var/lib/postgresql/data"
      ];
      service.healthcheck = {
        test = [ "CMD-SHELL" "pg_isready -U \${POSTGRES_USER:-certify}" ];
        interval = "10s";
        timeout = "5s";
        retries = 5;
      };
      service.restart = "unless-stopped";
    };

    redis = {
      service.image = "redis:7-alpine";
      service.container_name = "certify-redis";
      service.command = [
        "redis-server"
        "--maxmemory-policy"
        "allkeys-lru"
        "--maxmemory"
        "256mb"
      ];
      service.ports = [
        "\${REDIS_PORT:-6379}:6379"
      ];
      service.volumes = [
        "redis_data:/data"
      ];
      service.healthcheck = {
        test = [ "CMD" "redis-cli" "ping" ];
        interval = "10s";
        timeout = "5s";
        retries = 5;
      };
      service.restart = "unless-stopped";
    };
  };

  # Define named volumes
  docker-compose.raw = {
    volumes = {
      postgres_data = {
        driver = "local";
      };
      redis_data = {
        driver = "local";
      };
    };
  };
}
