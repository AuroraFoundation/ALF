{
  description = "Aurora Lyrics Format";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = {nixpkgs, ...}: let
    system = "x86_64-linux";
    pkgs = import nixpkgs {
      inherit system;
    };
  in
    with pkgs; {
      devShells."${system}".default = mkShell {
        packages = [
          gofumpt
          golangci-lint
          gotools
          gopls
          just
        ];

        buildInputs = [
          go_1_21
        ];

        shellHook = ''
          go version
        '';
      };
    };
}
