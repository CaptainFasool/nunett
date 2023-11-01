{ pkgs ? import <nixpkgs> {}
}:
{
  run-tx-validation-test = pkgs.writeScriptBin "run-tx-validation-test" ''
    #!${pkgs.runtimeShell}
    ${pkgs.go_1_20}/bin/go test -ldflags="-extldflags=-Wl,-z,lazy" -run=TestTheTestSuite
  '';
}
