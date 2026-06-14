# 🎉 Ready to Run! Pixel 9 API 37 Installed

## You Have Everything Ready ✅

- ✅ Pixel 9 emulator (API level 37)
- ✅ Android SDK installed
- ✅ Go backend running at http://localhost:8080

---

## 🚀 RUN THE APP NOW (2 Methods)

### Method 1: Using Android Studio (Easiest)

1. **Make sure backend is running** (in one terminal):
```bash
make run
```

2. **Start Android Studio**:
```bash
open -a "Android Studio" .
```

3. **In Android Studio**:
   - Wait for Gradle sync to complete
   - Click the green **Play** button (▶️) in the toolbar
   - Select **Pixel_9** from the device dropdown
   - Click **Run**
   - Wait ~30-60 seconds for first build
   - **App launches!** 🎉

### Method 2: Using Command Line (Faster)

```bash
# Terminal 1: Start backend
make run

# Terminal 2: Start emulator & run app
./START_APP.sh
```

This will:
- Start Pixel 9 emulator
- Wait for boot
- Build app
- Install on emulator
- Launch app

---

## 📊 What to Expect

### First Launch (30-60 seconds)
- Gradle downloads dependencies
- App compiles
- APK builds and installs
- App starts on emulator

### App UI
- **Home screen** - Welcome to Raaz
- **Matching screen** - "Enter Echo" button to join pool
- **Logs** - Should show WebSocket connection to `ws://10.0.2.2:8080/ws`

### Connection Flow
```
App → Emulator (10.0.2.2:8080) → Your Machine → Go Server (localhost:8080)
```

---

## 🐛 If Something Goes Wrong

### Check if emulator is running:
```bash
adb devices
# Should show: emulator-5554 device
```

### View app logs:
```bash
adb logcat | grep raaz
```

### Rebuild clean:
```bash
./gradlew clean build
./gradlew installDebug
adb shell am start -n com.raaz.app.debug/com.raaz.app.MainActivity
```

### Restart ADB:
```bash
adb kill-server
adb start-server
adb devices
```

---

## 💡 Common Commands

```bash
# List connected devices
adb devices

# View real-time logs
adb logcat | grep raaz

# Install app manually
./gradlew installDebug

# Start app manually
adb shell am start -n com.raaz.app.debug/com.raaz.app.MainActivity

# Stop emulator
adb emu kill

# Kill all adb
adb kill-server
```

---

## 📝 Checklist Before Running

- [ ] Backend running: `make run` (in Terminal 1)
- [ ] Emulator created: Pixel 9 API 37 ✅
- [ ] Android Studio installed ✅
- [ ] Project root open in Android Studio
- [ ] Gradle synced (wait for it to complete)

---

## 🎯 Your Next Step

**Choose one:**

**Option A (Recommended for beginners):**
```bash
open -a "Android Studio" .
# Click Play button
```

**Option B (For command line fans):**
```bash
./START_APP.sh
```

---

## 📞 Need Help?

**App won't start:**
```bash
adb logcat | grep raaz
# Paste the error here
```

**Can't connect to backend:**
- Verify `make run` is running in another terminal
- Check: `curl http://localhost:8080/health`

**Emulator won't boot:**
```bash
emulator -avd Pixel_9 -no-snapshot-load
```

---

**Ready? Let's go! 🚀**

Pick your method above and run it now!
