{
  description = "Personal GitHub repo config-as-code";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [ "x86_64-linux" ];

      imports = [ inputs.treefmt-nix.flakeModule ];

      perSystem =
        { pkgs, ... }:
        {
          devShells.default = pkgs.mkShell {
            packages = [
              pkgs.pulumi
              pkgs.pulumiPackages.pulumi-go
              pkgs.go
              pkgs.gopls
              pkgs.gh
            ];
          };

          treefmt = {
            projectRootFile = "flake.nix";
            programs = {
              nixfmt.enable = true;
              gofmt.enable = true;
              shellcheck.enable = true;
            };
          };
        };
    };
}
