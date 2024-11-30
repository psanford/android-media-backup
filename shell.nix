{ pkgs ? import <nixpkgs> {} }:

with import <nixpkgs> {
  config.android_sdk.accept_license = true;
  config.allowUnfree = true;
};

let
  pinPkgsFetch = pkgs.fetchFromGitHub {
    owner  = "NixOS";
    repo   = "nixpkgs";
    rev    = "31c5a59bf4fbb762bf763788f99c9d67d517d257";
    # Hash obtained using `nix-prefetch-url --unpack --unpack https://github.com/nixos/nixpkgs/archive/<rev>.tar.gz`
    sha256 = "0gs09fyqaj5mlbqh8k81pvc94nwgmnnjv07shdnvyi30mdapmji6";
  };
  pinPkgs = import pinPkgsFetch {
    config.android_sdk.accept_license = true;
    config.allowUnfree = true;
  };

  buildToolsVersion = "33.0.2";
  androidComposition = androidenv.composeAndroidPackages {
    platformVersions = [ "30" "33" ];
    buildToolsVersions = [ "${buildToolsVersion}" ];
    includeNDK = true;
  };
in

pkgs.mkShell {
  nativeBuildInputs = with pkgs.buildPackages; [
    openjdk17
    androidComposition.androidsdk
    pinPkgs.go_1_22
  ];

  shellHook = ''
    export GRADLE_OPTS="-Dorg.gradle.project.android.aapt2FromMavenOverride=${androidComposition.androidsdk}/libexec/android-sdk/build-tools/${buildToolsVersion}/aapt2";
    export ANDROID_SDK_ROOT="${androidComposition.androidsdk}/libexec/android-sdk"
  '';

}
