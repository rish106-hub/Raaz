# Raaz Android App - Local Setup Guide

## 📱 What You Have

- **Kotlin + Jetpack Compose** Android app
- **Firebase** integration (analytics, crashlytics)
- **WebSocket** client for real-time chat
- **Material Design 3** UI
- **Min SDK:** 26 (Android 8.0)
- **Target SDK:** 34 (Android 14)

---

## 🔧 Prerequisites

You need to install:

1. **Android Studio** (with SDK bundled)
2. **Java Development Kit (JDK)** 17+
3. **Android SDK** (API level 34)
4. **Emulator** or physical device

---

## 📥 Installation Steps

### Step 1: Install Android Studio

```bash
# Option A: Using Homebrew
brew install --cask android-studio

# Option B: Manual download
# Visit: https://developer.android.com/studio
# Download for macOS (Intel or Apple Silicon)
# Install the .dmg file
```

This installs:
- Android Studio IDE
- Bundled Android SDK
- Emulator
- ADB (Android Debug Bridge)

**Installation takes ~5-10 minutes**

### Step 2: Complete Android Studio Setup

1. **Open Android Studio** (first time will show setup wizard)
2. Accept default settings or customize:
   - Choose "Standard" installation
   - Install Android SDK 34
   - Install Google Play Services
   - Install Android Emulator
3. Click "Finish" and wait for downloads (~10-15 min)

### Step 3: Set Environment Variables

Add Android SDK to your PATH:

```bash
# Add to ~/.zshrc or ~/.bash_profile
export ANDROID_HOME=$HOME/Library/Android/sdk
export PATH=$ANDROID_HOME/emulator:$ANDROID_HOME/platform-tools:$PATH
```

Then reload:
```bash
source ~/.zshrc
```

Verify:
```bash
adb --version
emulator -version
```

### Step 4: Create or Connect Device

**Option A: Use Android Emulator (Easiest)**

```bash
# List available emulators
emulator -list-avds

# Create one if needed (from Android Studio)
# Tools > Device Manager > Create Device
# Select Pixel 6 or newer
# Select API 34 image

# Start emulator
emulator -avd Pixel_6_API_34
```

**Option B: Connect Physical Device**

```bash
# Enable Developer Mode on phone:
# Settings > About > Build number (tap 7 times)
# Settings > Developer options > USB Debugging (enable)

# Connect with USB cable
adb devices
# Should show your device
```

---

## 🚀 Running the App

### Step 1: Ensure Backend is Running

```bash
# In another terminal, start the Go server
make run
# Or: ./raaz-server
# Should show: raaz server starting port=8080
```

### Step 2: Open Project in Android Studio

```bash
# From project root
open -a "Android Studio" .
```

Or manually:
1. Open Android Studio
2. File → Open → Select `/path/to/raaz` folder
3. Wait for Gradle build (takes 2-3 min first time)

### Step 3: Build & Run

**From Android Studio:**

1. Click the green **Play** button (▶️) in toolbar
2. Select device or emulator
3. Click "Run"
4. Wait for build (~30-60 seconds)
5. App launches on device/emulator

**From Command Line:**

```bash
./gradlew installDebug
adb shell am start -n com.raaz.app.debug/com.raaz.app.MainActivity
```

---

## 📋 Project Structure

```
app/
├── src/main/
│   ├── AndroidManifest.xml        # App permissions & activities
│   ├── java/com/raaz/app/
│   │   ├── MainActivity.kt        # Entry point
│   │   ├── RaazApplication.kt     # App initialization
│   │   ├── RaazApp.kt            # Compose root
│   │   ├── features/
│   │   │   ├── home/             # Home screen
│   │   │   ├── matching/         # Matching flow
│   │   │   ├── chat/             # Chat screen
│   │   │   └── profile/          # User profile
│   │   └── data/
│   │       ├── repository/       # Business logic
│   │       ├── local/            # Local storage (Room DB)
│   │       └── remote/           # API calls (WebSocket)
│   └── res/
│       ├── values/               # Strings, colors, dimensions
│       ├── drawable/             # Images, icons
│       └── layout/               # Legacy layouts
├── build.gradle.kts              # App-level Gradle config
└── proguard-rules.pro            # Code obfuscation
```

---

## 🔌 WebSocket Configuration

The app connects to your backend server. Configuration by build type:

### Debug Build (Development)
```
WS_BASE_URL = "ws://10.0.2.2:8080/ws"
```
- `10.0.2.2` = Special hostname that resolves to localhost when running in emulator
- Physical devices use `ws://192.168.x.x:8080/ws` (your machine's IP)

### Release Build (Production)
```
WS_BASE_URL = "wss://api.raaz.app/ws"
```
- Uses production server with SSL

### Change for Local Testing

If using physical device, update `app/build.gradle.kts`:

```kotlin
debug {
    buildConfigField("String", "WS_BASE_URL", "\"ws://YOUR_IP:8080/ws\"")
}
```

Get your IP:
```bash
ifconfig | grep "inet " | grep -v 127.0.0.1
# Use the 192.168.x.x address
```

---

## 🧪 Testing

### Run Unit Tests
```bash
./gradlew test
```

### Run Instrumentation Tests (on device/emulator)
```bash
./gradlew connectedAndroidTest
```

### View Logs
```bash
adb logcat | grep raaz
```

---

## 🐛 Troubleshooting

| Issue | Solution |
|-------|----------|
| "SDK not found" | Install Android Studio (includes SDK) |
| Emulator won't start | Enable virtualization in BIOS, or use physical device |
| App crashes on startup | Check logs: `adb logcat` |
| Can't connect to backend | Ensure Go server running, check WebSocket URL in logs |
| Gradle build fails | Run `./gradlew clean build` |
| Firebase error | Copy `google-services.json.example` → `google-services.json` |

---

## 📚 Key Files to Understand

| File | Purpose |
|------|---------|
| `MainActivity.kt` | Activity entry point |
| `RaazApp.kt` | Compose UI root |
| `MatchingViewModel.kt` | Matching logic & state |
| `HomeViewModel.kt` | Home screen logic |
| `build.gradle.kts` | Dependencies & build config |

---

## 💡 Development Tips

```bash
# Clear build cache
./gradlew clean

# Build without running
./gradlew build

# Check dependencies
./gradlew dependencies

# Format code
./gradlew spotlessApply

# Run specific test
./gradlew test --tests com.raaz.app.*Test
```

---

## 🎯 First Time Checklist

- [ ] Install Android Studio
- [ ] Complete Android Studio setup wizard
- [ ] Set ANDROID_HOME environment variable
- [ ] Verify: `adb --version` works
- [ ] Create emulator or connect physical device
- [ ] Start Go backend: `make run`
- [ ] Open project in Android Studio
- [ ] Click Play button to run app
- [ ] App launches on device!

---

## 📞 Getting Help

**Android Logcat (most important for debugging)**
```bash
adb logcat | grep raaz
```

**Check emulator network**
```bash
adb shell ping 10.0.2.2
```

**Stop all adb services**
```bash
adb kill-server
adb start-server
```

---

**Next Step:** Install Android Studio and come back! 🚀
