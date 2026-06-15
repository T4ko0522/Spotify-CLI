{
  lib,
  buildGo125Module,
}:

buildGo125Module rec {
  pname = "spotify-cli";
  version = "1.3.0";

  src = ./.;

  vendorHash = "sha256-nPfgI+hSA1pOWrHZN8qJPlA0k1OdtF50n+3h5zCA3q4=";

  ldflags = [
    "-s"
    "-w"
    "-X main.version=v${version}"
  ];

  postInstall = ''
    if [ -e "$out/bin/spotify-cli" ]; then
      mv "$out/bin/spotify-cli" "$out/bin/spt"
    fi
  '';

  doInstallCheck = true;
  installCheckPhase = ''
    runHook preInstallCheck

    $out/bin/spt --help >/dev/null
    $out/bin/spt --version | grep "v${version}" >/dev/null

    runHook postInstallCheck
  '';

  meta = {
    description = "Control Spotify from your terminal with a TUI and CLI";
    homepage = "https://github.com/T4ko0522/spotify-cli";
    license = lib.licenses.asl20;
    maintainers = with lib.maintainers; [ ];
    mainProgram = "spt";
    platforms = lib.platforms.linux;
  };
}
