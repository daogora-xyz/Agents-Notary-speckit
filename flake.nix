{
  description = "Agents Notary spec-kit development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    arion.url = "github:hercules-ci/arion";
  };

  outputs = { self, nixpkgs, flake-utils, arion }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };

        # Build spec-kit (specify-cli) from GitHub
        spec-kit = pkgs.python311Packages.buildPythonApplication {
          pname = "specify-cli";
          version = "unstable-2025-01-15";

          src = pkgs.fetchFromGitHub {
            owner = "github";
            repo = "spec-kit";
            rev = "main";
            sha256 = "sha256-A5WQ6/YeEfYrGRxO/V7grKB3O2wv4WIXBvNBAYxAx4Y=";
          };

          format = "pyproject";

          nativeBuildInputs = with pkgs.python311Packages; [
            hatchling
            wheel
          ];

          propagatedBuildInputs = with pkgs.python311Packages; [
            click
            rich
            pyyaml
            jinja2
            httpx
            platformdirs
            readchar
            truststore
            typer
          ];

          # Don't check since we need to update the hash first
          doCheck = false;

          meta = with pkgs.lib; {
            description = "GitHub's toolkit for Spec-Driven Development";
            homepage = "https://github.com/github/spec-kit";
            license = licenses.mit;
            mainProgram = "specify";
          };
        };

        # Arion configuration
        arionPkgs = import arion.inputs.nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = [
            # spec-kit
            spec-kit

            # Development tools for the project
            pkgs.go
            pkgs.gopls
            pkgs.gotools
            pkgs.go-tools

            # Database and infrastructure (native for tooling)
            pkgs.postgresql_16
            pkgs.redis

            # Arion for Docker Compose via Nix
            arion.packages.${system}.arion

            # Build tools
            pkgs.git
            pkgs.gnumake

            # Container runtime (Podman - daemonless alternative to Docker)
            pkgs.podman
            pkgs.podman-compose

            # Docker tools (for compatibility, but prefer Podman)
            pkgs.docker-client
            pkgs.docker-compose

            # Utilities
            pkgs.jq
            pkgs.curl
            pkgs.direnv
          ];

          shellHook = ''
            echo "🚀 Agents Notary spec-kit development environment loaded!"
            echo ""
            echo "Available tools:"
            echo "  - spec-kit: Kiro Spec-Driven Development toolkit"
            echo "  - go ($(go version | cut -d' ' -f3))"
            echo "  - podman: Daemonless container runtime"
            echo "  - arion: NixOS Docker Compose (works with Podman)"
            echo "  - postgresql ${pkgs.postgresql_16.version} (native tooling)"
            echo "  - redis ${pkgs.redis.version} (native tooling)"
            echo ""
            echo "Project: certify.ar4s.com - Blockchain Certification Platform"
            echo "Spec: docs/OVERVIEW.md"
            echo ""
            echo "Container Services:"
            echo "  - With Podman: podman-compose up -d"
            echo "  - With Arion:  arion up -d (uses Podman backend)"
            echo "  - Stop:        podman-compose down (or arion down)"
            echo "  - Logs:        podman-compose logs -f"
            echo ""
            echo "💡 Tip: Podman is daemonless and works without root!"
            echo ""
          '';

          # Configure Podman to use Docker socket compatibility
          DOCKER_HOST = "unix:///run/user/\${UID}/podman/podman.sock";

          # Environment variables
          PGDATA = "$PWD/.postgres";
          REDIS_DATA = "$PWD/.redis";
        };

        # Optional: Add a package output if you want to build something
        packages.default = pkgs.writeShellScriptBin "certify-dev" ''
          echo "Certify.ar4s.com development environment"
          exec ${pkgs.bashInteractive}/bin/bash
        '';
      }
    );
}
