# Verifikasi Lingkungan Modul 10 — SIAKAD Mini

Dokumen ini mencatat hasil verifikasi lingkungan kerja untuk pengerjaan Modul 10 (API dengan MongoDB).

## 1. Komponen & Versi Terpasang

| Komponen | Status | Versi / Detail | Keterangan |
|---|---|---|---|
| **Sistem Operasi** | Aktif | Windows 11 (amd64) | Host development |
| **Go** | Terpasang | `go version go1.26.5 windows/amd64` | Memenuhi syarat minimal Go 1.25+ |
| **PostgreSQL** | Berjalan | PostgreSQL 18 (Port 5432) | Database `siakad_mini` dengan migrasi 001 - 006 |
| **MongoDB Service** | Running | Service Windows: `MongoDB` | Dijalankan secara otomatis sebagai Windows Service |
| **MongoDB Server** | Terpasang | `v8.3.7` | `C:\Program Files\MongoDB\Server\8.3\bin\mongod.exe` |
| **mongosh (Shell)** | Terpasang | `v2.14.0` | `C:\Program Files\mongosh\mongosh.exe` |
| **GUI Tool** | Terpasang | TablePlus | `C:\Program Files\TablePlus` (digunakan sebagai GUI lokal pengganti MongoDB Atlas / Compass) |

---

## 2. Hasil Verifikasi Service & Koneksi

### 2.1 Verifikasi Service MongoDB
Perintah:
```powershell
Get-Service MongoDB
```

Hasil:
```text
Status   Name               DisplayName
------   ----               -----------
Running  MongoDB            MongoDB Server (MongoDB)
```

### 2.2 Verifikasi Versi MongoDB Server
Perintah:
```powershell
mongod --version
```

Hasil:
```text
db version v8.3.7
Build Info: {
    "version": "8.3.7",
    "gitVersion": "34eee04f34989abb7a3d91447976f033f4f74af2",
    "modules": [],
    "allocator": "tcmalloc-gperf",
    "environment": {
        "distmod": "windows",
        "distarch": "amd64",
        "target_arch": "amd64"
    }
}
```

### 2.3 Verifikasi Versi `mongosh`
Perintah:
```powershell
mongosh --version
```

Hasil:
```text
2.14.0
```

### 2.4 Uji Ping Koneksi ke MongoDB
Perintah:
```powershell
mongosh --eval "db.runCommand({ ping: 1 })"
```

Hasil:
```json
{ "ok": 1 }
```
Koneksi ke instance lokal MongoDB berhasil terverifikasi.

