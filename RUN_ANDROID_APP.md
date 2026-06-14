# 🚀 How to Run Raaz Android App Locally

## Current Status

✅ **Go Backend**: Running at http://localhost:8080  
❌ **Android App**: Not installed yet (needs Android Studio)

---

## 3-Step Quick Setup

### Step 1️⃣: Install Android Studio (10 min)

```bash
brew install --cask android-studio
```

Or download manually from: https://developer.android.com/studio

**What it includes:**
- Android SDK (includes API 34, emulator, etc)
- Gradle build system
- ADB (Android Debug Bridge)
- Emulator (to test on virtual device)

### Step 2️⃣: Create/Start Emulator

**Option A: Use Emulator (easiest)**
```bash
# Open Android Studio
open -a "Android Studio" .

# In Android Studio:
# 1. Tools → Device Manager
# 2. Click "Create Device"
# 3. Select "Pixel 6" or any recent phone
# 4. Select "API 34 (Android 14)"
# 5. Click "Finish"
# 6. Click the play icon next to your device to start
```

**Option B: Use Physical Phone**
```bash
# Connect phone via USB cable
# On phone: Settings → About → tap "Build number" 7 times
# Settings → Developer options → USB Debugging (enable)
# Confirm USB authorization prompt on phone
adb devices  # Should show your device
```

### Step 3️⃣: Run the App

**Make sure Go backend is running:**
```bash
# Terminal 1: Backend
make run
```

**Build and run Android app:**
```bash
# Terminal 2: From project root
open -a "Android Studio" .
```

Then in Android Studio:
1. Click green **Play** button (▶️) top toolbar
2. Select emulator or physical device
3. Click "Run"
4. Wait ~30 seconds for build
5. **App launches!** 🎉

---

## What to Expect

### First Launch
- App shows Raaz home screen
- May show loading spinner while connecting to backend
- If backend is running, should connect successfully

### Features to Test
- ✅ Open home screen
- ✅ Go to "Matching" screen
- ✅ Tap "Enter Echo" to join matching pool
- ✅ See real-time connection status
- ✅ (Need 2nd device/emulator to actually match)

### If App Crashes
Check logs in Android Studio:
```bash
# Terminal
adb logcat | grep -i raaz

# Or in Android Studio:
# View → Tool Windows → Logcat
```

---

## Detailed Setup Guide

See **ANDROID_SETUP.md** for:
- Detailed installation steps with screenshots
- Environment configuration
- WebSocket configuration
- Troubleshooting guide
- Development tips

---

## File Structure

```
project/
├── Backend (Go)
│   ├── server/main.go
│   ├── server/app.go
│   ├── raaz-server (binary)
│   └── Makefile
│
└── Frontend (Android/Kotlin)
    ├── app/build.gradle.kts
    ├── app/src/main/
    │   ├── AndroidManifest.xml
    │   └── java/com/raaz/app/
    │       ├── MainActivity.kt
    │       ├── RaazApp.kt
    │       ├── features/
    │       │   ├── home/
    │       │   ├── matching/
    │       │   └── chat/
    │       └── data/
    ├── gradle/
    └── settings.gradle.kts
```

---

## Common Commands

```bash
# Build Android app
./gradlew build

# Build and install on device
./gradlew installDebug

# Start app on device
adb shell am start -n com.raaz.app.debug/com.raaz.app.MainActivity

# View logs
adb logcat

# Stop adb daemon (fix connection issues)
adb kill-server
adb start-server

# List connected devices
adb devices
```

---

## Next Steps After Installation

1. ✅ Install Android Studio
2. ✅ Create emulator
3. ✅ Run backend: `make run`
4. ✅ Run app from Android Studio
5. 📖 Read the code in `app/src/main/java/com/raaz/app/`
6. 🔍 Test WebSocket connection in logs
7. 🎨 Modify UI (it's Jetpack Compose - hot reload works!)

---

## WebSocket Connection

The app auto-connects to your backend on launch.

**Debug build uses:** `ws://10.0.2.2:8080/ws`
- `10.0.2.2` = localhost when running in Android emulator
- Physical device needs your machine's IP (e.g., `192.168.1.x`)

To change for physical device:
1. Get your IP: `ifconfig | grep "inet " | grep -v 127.0.0.1`
2. Edit `app/build.gradle.kts`
3. Change `WS_BASE_URL` in debug block
4. Rebuild app

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "Cannot connect to backend" | Ensure `make run` is executing in another terminal |
| App crashes immediately | Check `adb logcat` for errors |
| Can't find emulator | Run `emulator -list-avds` to list, or create new one |
| Gradle build hangs | Run `./gradlew clean build` |
| "Build-tools not found" | Open Android Studio and complete setup wizard |

---

## Resources

- 📱 **Android Docs**: https://developer.android.com/docs
- 🎨 **Jetpack Compose**: https://developer.android.com/jetpack/compose
- 🔌 **WebSocket (OkHttp)**: https://square.github.io/okhttp/
- 🐛 **Android Logcat**: https://developer.android.com/studio/debug/logcat

---

**Ready?** Start with `brew install --cask android-studio` 🚀
