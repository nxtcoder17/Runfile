{
  description = "RunFile dev workspace";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          # hardeningDisable = [ "all" ];

          buildInputs = with pkgs; [
            # cli tools
            curl
            jq
            yq

            pre-commit

            # programming tools
            go_1_24

            upx

            gotestfmt
          ];

          shellHook = ''
          '';
        };
      }
    );
}


