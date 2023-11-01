{ pkgs ? import <nixpkgs> {}
, system ? builtins.currentSystem
}:

let
  def = import ./default.nix { inherit pkgs; };
in

with pkgs;

mkShell {
  buildInputs = [
    def.run-tx-validation-test
    systemd
  ];
}
