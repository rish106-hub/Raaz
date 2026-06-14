#!/bin/bash

echo "╔════════════════════════════════════════════════════════╗"
echo "║  Raaz Android App - Quick Setup                        ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Check if Android Studio is installed
if ! command -v adb &> /dev/null; then
    echo "❌ Android SDK not found"
    echo ""
    echo "📥 STEP 1: Install Android Studio"
    echo ""
    echo "   Option A (Recommended - Homebrew):"
    echo "   $ brew install --cask android-studio"
    echo ""
    echo "   Option B (Manual):"
    echo "   → Visit https://developer.android.com/studio"
    echo "   → Download for macOS"
    echo "   → Install the .dmg file"
    echo ""
    echo "📖 Full guide: See ANDROID_SETUP.md"
    exit 1
fi

echo "✅ Android SDK found: $(adb --version | head -1)"
echo ""

# List emulators
echo "📱 Available emulators:"
if command -v emulator &> /dev/null; then
    emulator -list-avds || echo "   (none - create one in Android Studio)"
else
    echo "   (emulator not in PATH - set ANDROID_HOME)"
fi

echo ""
echo "═════════════════════════════════════════════════════════"
echo ""
echo "🚀 TO RUN THE APP:"
echo ""
echo "1. Start the Go backend (in another terminal):"
echo "   $ make run"
echo ""
echo "2. Open Android Studio:"
echo "   $ open -a \"Android Studio\" ."
echo ""
echo "3. Create/start emulator (if needed):"
echo "   Tools → Device Manager → Create Device"
echo ""
echo "4. Click Play button (▶️) to run app"
echo ""
echo "═════════════════════════════════════════════════════════"
