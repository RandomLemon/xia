{
  description = "xia — kei + chat-agent chatbot host";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
    in
    {
      devShells = nixpkgs.lib.genAttrs systems (
        system:
        {
          default = import ./nix/devshell.nix {
            pkgs = nixpkgs.legacyPackages.${system};
          };
        }
      );
    };
}
