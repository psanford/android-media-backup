{
  description = "Android Media Backup - Go/Android development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config = {
            android_sdk.accept_license = true;
            allowUnfree = true;
          };
        };

        buildToolsVersion = "35.0.0";
        androidComposition = pkgs.androidenv.composeAndroidPackages {
          platformVersions = [ "35" ];
          buildToolsVersions = [ buildToolsVersion ];
          includeNDK = true;
          ndkVersions = [ "27.2.12479018" ];
        };

        androidSdk = androidComposition.androidsdk;
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go
            go_1_25
            gomobile

            # Java
            openjdk21
            gradle

            # Android SDK
            androidSdk

            # Build tools
            gnumake
          ];

          shellHook = ''
            export ANDROID_SDK_ROOT="${androidSdk}/libexec/android-sdk"
            export ANDROID_NDK_ROOT="${androidSdk}/libexec/android-sdk/ndk/27.2.12479018"
            export GRADLE_OPTS="-Dorg.gradle.project.android.aapt2FromMavenOverride=${androidSdk}/libexec/android-sdk/build-tools/${buildToolsVersion}/aapt2"

            echo "Android Media Backup development environment"
            echo "  Go:          $(go version)"
            echo "  Java:        $(java -version 2>&1 | head -1)"
            echo "  Android SDK: $ANDROID_SDK_ROOT"
            echo "  Android NDK: $ANDROID_NDK_ROOT"
          '';
        };
      }
    );
}
