{
  description = "Lightweight TUI to watch tmux sessions";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    in
    {
      packages = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          tmuxwatch = pkgs.buildGoModule {
            pname = "tmuxwatch";
            version = "0.9.1";

            src = ./.;

            vendorHash = "sha256-hNszgsQ4lZ6NinZs3IiPnYf3Jl/eSSLyBSw2wiLqyDc=";

            subPackages = [ "cmd/tmuxwatch" ];

            excludedPackages = [ "tools" ];

            meta = with pkgs.lib; {
              description = "Lightweight TUI to watch tmux sessions";
              homepage = "https://github.com/steipete/tmuxwatch";
              license = licenses.mit;
              maintainers = [ ];
              mainProgram = "tmuxwatch";
            };
          };

          default = self.packages.${system}.tmuxwatch;
        }
      );

      apps = forAllSystems (system: {
        tmuxwatch = {
          type = "app";
          program = "${self.packages.${system}.tmuxwatch}/bin/tmuxwatch";
        };

        default = self.apps.${system}.tmuxwatch;
      });

      devShells = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
              golangci-lint
              gofumpt
              tmux
            ];
          };
        }
      );
    };
}
