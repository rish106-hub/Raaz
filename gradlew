#!/bin/bash
cd "$(dirname "$0")"
GRADLE_VERSION="8.10"
GRADLE_USER_HOME="${GRADLE_USER_HOME:-$HOME/.gradle}"
GRADLE_HOME="$GRADLE_USER_HOME/wrapper/dists/gradle-${GRADLE_VERSION}-bin"

# Find the actual gradle installation
if [ -d "$GRADLE_HOME" ]; then
  # Look for gradle-8.10 subdirectory in hash directories
  GRADLE_BIN=$(find "$GRADLE_HOME" -name "gradle" -type f | head -1)
  if [ -n "$GRADLE_BIN" ]; then
    "$GRADLE_BIN" "$@"
    exit $?
  fi
fi

# Download gradle if not found
mkdir -p "$(dirname "$GRADLE_HOME")"
echo "Downloading Gradle $GRADLE_VERSION..."
TMP_ZIP="/tmp/gradle-$GRADLE_VERSION-bin.zip"
curl -L -o "$TMP_ZIP" "https://services.gradle.org/distributions/gradle-$GRADLE_VERSION-bin.zip" || exit 1
unzip -q -d "$GRADLE_HOME" "$TMP_ZIP" || exit 1
rm "$TMP_ZIP"

# Find and run gradle
GRADLE_BIN=$(find "$GRADLE_HOME" -name "gradle" -type f | head -1)
if [ -n "$GRADLE_BIN" ]; then
  "$GRADLE_BIN" "$@"
else
  echo "Gradle not found after extraction."
  exit 1
fi
