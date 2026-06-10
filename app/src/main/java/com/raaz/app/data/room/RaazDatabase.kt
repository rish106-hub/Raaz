package com.raaz.app.data.room

import android.content.Context
import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase

@Database(
    entities = [ChatSession::class, VaultMessage::class],
    version = 1,
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
                ).build().also { INSTANCE = it }
            }
        }
    }
}
