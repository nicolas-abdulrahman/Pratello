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
          buildInputs = with pkgs; [
            go
            nodejs_22
            docker
            docker-compose
            git
          ];

          shellHook = ''
            echo "Go: $(go version)"
            echo "Node: $(node --version)"
            echo "Git: $(git --version)"
          '';
        };
      });
}
