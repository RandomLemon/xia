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
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            name = "xia";

            packages = with pkgs; [
              go
              gcc
              gopls
              gotools
              golangci-lint
              git
              curl
              jq
            ];

            env.GOTOOLCHAIN = "local";

            shellHook = ''
              echo "xia dev shell · $(go version)"
              echo "llmstub: go run ./cmd/llmstub"
              echo "chatbot: go run ./cmd/xia -config configs/config.yaml"
            '';
          };
        }
      );
    };
}
