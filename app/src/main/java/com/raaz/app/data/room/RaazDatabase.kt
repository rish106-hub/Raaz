package com.raaz.app.data.room

import android.content.Context
import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase
import androidx.room.migration.Migration
import androidx.sqlite.db.SupportSQLiteDatabase

// M-14: bump to version 2; adds extensionsUsed column to chat_sessions
val MIGRATION_1_2 = object : Migration(1, 2) {
    override fun migrate(database: SupportSQLiteDatabase) {
        database.execSQL(
            "ALTER TABLE chat_sessions ADD COLUMN extensionsUsed INTEGER NOT NULL DEFAULT 0"
        )
    }
}

@Database(
    entities = [ChatSession::class, VaultMessage::class],
    version = 2,
    exportSchema = false
)
abstract class RaazDatabase : RoomDatabase() {
    abstract fun chatSessionDao(): ChatSessionDao
    abstract fun vaultDao(): VaultDao

    companion object {
        @Volatile
        private var INSTANCE: RaazDatabase? = null

        fun getInstance(context: Context): RaazDatabase {
            return INSTANCE ?: synchronized(this) {
                INSTANCE ?: Room.databaseBuilder(
                    context.applicationContext,
                    RaazDatabase::class.java,
                    "raaz_database"
                )
                    .addMigrations(MIGRATION_1_2)
                    .build()
                    .also { INSTANCE = it }
            }
        }
    }
}
