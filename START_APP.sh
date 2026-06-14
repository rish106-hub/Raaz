#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════╗"
echo "║  Raaz - Start Emulator & Run App                       ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Ensure Android paths are set
export ANDROID_HOME=$HOME/Library/Android/sdk
export PATH=$ANDROID_HOME/emulator:$ANDROID_HOME/platform-tools:$PATH

# Step 1: Start emulator
echo "🚀 STEP 1: Starting Pixel 9 emulator..."
echo "   (This takes ~30-60 seconds)"
echo ""
emulator -avd Pixel_9 -no-snapshot-load -no-boot-anim &
EMULATOR_PID=$!

# Wait for emulator to boot
echo "⏳ Waiting for emulator to boot..."
sleep 15

# Check if adb recognizes device
for i in {1..30}; do
  if adb devices | grep -q "device$"; then
    echo "✅ Emulator booted successfully"
    break
  fi
  if [ $i -eq 30 ]; then
    echo "❌ Emulator failed to boot"
    exit 1
  fi
  sleep 2
done

echo ""
echo "════════════════════════════════════════════════════════"
echo ""
echo "📱 STEP 2: Build & Run App"
echo ""
echo "🔨 Building app (first build takes ~30-60 seconds)..."
cd "$(dirname "$0")"
./gradlew installDebug

echo ""
echo "📱 Launching app on emulator..."
adb shell am start -n com.raaz.app.debug/com.raaz.app.MainActivity

echo ""
echo "════════════════════════════════════════════════════════"
echo ""
echo "✅ App launched on emulator!"
echo ""
echo "💡 TIPS:"
echo "   • View logs: adb logcat | grep raaz"
echo "   • Stop emulator: kill $EMULATOR_PID"
echo "   • Kill adb: adb kill-server"
echo ""
echo "🌐 Backend should be running:"
echo "   $ make run (in another terminal)"
echo ""
