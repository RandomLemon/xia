# devShell 的唯一实现：flake（nix develop）与 shell.nix（nix-shell）共用，
# 避免两套入口的工具链漂移。
{ pkgs }:
pkgs.mkShell {
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

  # 只使用 devShell 提供的 Go，禁止自动下载其它 toolchain。
  env.GOTOOLCHAIN = "local";

  shellHook = ''
    echo "xia dev shell · $(go version)"
    echo "llmstub: go run ./cmd/llmstub"
    echo "chatbot: go run ./cmd/xia -config configs/config.yaml"
  '';
}
