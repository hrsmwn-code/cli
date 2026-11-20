{
  description = "pangolin-cli - a VPN client for pangolin";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    # The official modules lack the TCP fallback code. Pin the fork sources
    # for both vendoring and compilation without changing the Go release workflow.
    newtFork = {
      url = "github:hrsmwn-code/newt/38e8716b7f12e9e37b6c534d731d35294c2ce33c";
      flake = false;
    };
    olmFork = {
      url = "github:hrsmwn-code/olm/b81c371c00dbbb67cf0c4fe47b55ab65ade0c076";
      flake = false;
    };
  };

  outputs = {nixpkgs, newtFork, olmFork, ...}: let
    supportedSystems = [
      "x86_64-linux"
      "aarch64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
    ];
    forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    pkgsFor = system: nixpkgs.legacyPackages.${system};
  in {
    packages = forAllSystems (
      system: let
        pkgs = pkgsFor system;
      in rec {
        pangolin-cli = pkgs.buildGoModule (finalAttrs: {
          pname = "pangolin-cli";
          version = "0.23.0-fallback";
          src = ./.;
          subPackages = [ "." ];

          vendorHash = "sha256-Llgk+nQNCEC1dhr02y+Hf3az8tmfxdhZX7c3SDzZ9N4=";

          postPatch = ''
            mkdir -p .fork-deps
            ln -s ${newtFork} .fork-deps/newt
            ln -s ${olmFork} .fork-deps/olm
            go mod edit \
              -replace github.com/fosrl/newt=./.fork-deps/newt \
              -replace github.com/fosrl/olm=./.fork-deps/olm
          '';

          ldflags = [
            "-s"
            "-w"
            "-X github.com/fosrl/cli/internal/version.Version=${finalAttrs.version}"
            "-X github.com/fosrl/cli/internal/version.NewtVersionOverride=1.20.1"
            "-X github.com/fosrl/cli/internal/version.OlmVersionOverride=1.10.2"
          ];

          postInstall = ''
            mv "$out/bin/cli" "$out/bin/pangolin-cli"
          '';

          meta = with pkgs.lib; {
            description = "A VPN client for pangolin";
            homepage = "https://github.com/fosrl/cli";
            license = licenses.unfree;
            maintainers = [];
            mainProgram = "pangolin-cli";
          };
        });

        default = pangolin-cli;
      }
    );

    devShells = forAllSystems (
      system: let
        pkgs = pkgsFor system;

        inherit
          (pkgs)
          go
          golangci-lint
          ;
      in {
        default = pkgs.mkShell {
          buildInputs = [
            go
            golangci-lint
          ];
        };
      }
    );
  };
}
