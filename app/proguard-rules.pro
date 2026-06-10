# OkHttp
-dontwarn okhttp3.**
-dontwarn okio.**
-keep class okhttp3.** { *; }
-keep interface okhttp3.** { *; }
-keepattributes Signature
-keepattributes *Annotation*

# Gson
-keep class com.google.gson.** { *; }
-keep class * implements com.google.gson.TypeAdapterFactory
-keep class * implements com.google.gson.JsonSerializer
-keep class * implements com.google.gson.JsonDeserializer
-keepclassmembers,allowobfuscation class * {
    @com.google.gson.annotations.SerializedName <fields>;
}

# Room — keep entities and DAOs so generated code resolves at runtime
-keep class * extends androidx.room.RoomDatabase { *; }
-keep @androidx.room.Entity class * { *; }
-keep @androidx.room.Dao class * { *; }
-keepclassmembers class * extends androidx.room.RoomDatabase {
    abstract *;
}

# WorkManager — keep Worker subclasses for reflection-based instantiation
-keep class * extends androidx.work.Worker { *; }
-keep class * extends androidx.work.CoroutineWorker { *; }
-keepclassmembers class * extends androidx.work.Worker {
    public <init>(android.content.Context, androidx.work.WorkerParameters);
}
-keepclassmembers class * extends androidx.work.CoroutineWorker {
    public <init>(android.content.Context, androidx.work.WorkerParameters);
}

# Kotlin coroutines
-keepclassmembernames class kotlinx.** {
    volatile <fields>;
}
-keepclassmembers class kotlin.coroutines.** { *; }
-dontwarn kotlinx.coroutines.**

# App data models — kept for Room and Gson reflection
-keep class com.raaz.app.data.** { *; }
-keep class com.raaz.app.data.websocket.** { *; }

# Compose — keep lambdas stable for recomposition tracking
-keepclassmembers class * implements androidx.compose.runtime.Stable { *; }

# General Kotlin
-keep class kotlin.** { *; }
-keep class kotlin.Metadata { *; }
-dontwarn kotlin.**
-keepclassmembers class **$WhenMappings {
    <fields>;
}
-keepclassmembers class kotlin.Lazy {
    <fields>;
}
