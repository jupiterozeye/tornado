{
  description = "Tornado - terminal SQL client";

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
        packages.tornado = pkgs.buildGoModule {
          pname = "tornado";
          version = "dev";
          src = ./.;
          subPackages = [ "cmd/tornado" ];
          vendorHash = "sha256-sibMXUN+NCuDhEfFv1Vh2/qy4JUEvEY2YFEQVGMAKl8=";

          meta = with pkgs.lib; {
            description = "Keyboard-first TUI SQL client";
            homepage = "https://github.com/jupiterozeye/tornado";
            license = licenses.mit;
            mainProgram = "tornado";
          };
        };

        packages.default = self.packages.${system}.tornado;

        apps.default = {
          type = "app";
          program = "${self.packages.${system}.tornado}/bin/tornado";
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools
            golangci-lint
          ];
        };
      });
}
