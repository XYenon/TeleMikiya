_:

{
  projectRootFile = "flake.nix";
  programs = {
    deadnix.enable = true;
    gofmt.enable = true;
    goimports.enable = true;
    nixfmt.enable = true;
    statix.enable = true;
    taplo.enable = true;
    yamlfmt.enable = true;
  };
}
