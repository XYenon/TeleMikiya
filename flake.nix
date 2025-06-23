{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    nur = {
      url = "github:nix-community/NUR";
      inputs = {
        nixpkgs.follows = "nixpkgs";
        treefmt-nix.follows = "treefmt-nix";
      };
    };
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      treefmt-nix,
      ...
    }@inputs:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = with inputs; [ nur.overlays.default ];
        };
        treefmtEval = treefmt-nix.lib.evalModule pkgs ./treefmt.nix;
      in
      {
        checks.formatting = treefmtEval.config.build.check self;
        formatter = treefmtEval.config.build.wrapper;
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gotools
            nur.repos.xyenon.go-check
          ];
        };
        packages.default =
          let
            rev = self.shortRev or self.dirtyShortRev or "dirty";
            date = self.lastModifiedDate or self.lastModified or "19700101";
            version = "${builtins.substring 0 8 date}_${rev}";
          in
          pkgs.buildGoModule {
            pname = "telemikiya";
            inherit version;
            src = ./.;
            vendorHash = "sha256-pMhdFs8KVqULfG3Ry8v1/o1tgflpxcw2ntJXE92ep/s=";
            subPackages = [ "." ];
          };
      }
    );
}
