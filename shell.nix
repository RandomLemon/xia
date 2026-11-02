# 传统 nix-shell 入口（不需要 flakes）：在仓库根目录执行 `nix-shell` 即可。
# nixpkgs 版本直接读 flake.lock，保证与 `nix develop` 使用同一份锁定。
let
  lock = builtins.fromJSON (builtins.readFile ./flake.lock);
  pinned = lock.nodes.nixpkgs.locked;
  nixpkgs = builtins.fetchTarball {
    url = "https://github.com/${pinned.owner}/${pinned.repo}/archive/${pinned.rev}.tar.gz";
    sha256 = pinned.narHash;
  };
  pkgs = import nixpkgs { system = builtins.currentSystem; };
in
import ./nix/devshell.nix { inherit pkgs; }
