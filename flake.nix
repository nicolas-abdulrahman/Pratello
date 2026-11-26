{
  description = "Pratello - Svelte + Go + Tailwind project environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          DATABASE_URL = "postgres://nick@localhost:5432/pratello?sslmode=disable";

          buildInputs = with pkgs; [
            go
            nodejs_22
            docker
            docker-compose
            git
            postgresql
          ];

          shellHook = ''
            export DATABASE_URL="postgres://nick@localhost:5432/pratello?sslmode=disable"
            echo "DATABASE_URL: $DATABASE_URL"
            echo "Go: $(go version)"
            echo "Node: $(node --version)"
            echo "Git: $(git --version)"
          '';
        };
      });
}
