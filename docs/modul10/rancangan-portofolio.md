# Rancangan Dokumen Portofolio Mahasiswa

Collection: `student_portfolios`  
Database: `siakad_mini` (MongoDB)

Dokumen ini dirancang dengan pendekatan *embedding* (dokumen bersarang) untuk data `projects` dan `certificates`. Karena proyek dan sertifikat merupakan milik eksklusif dari mahasiswa yang bersangkutan dan selalu dibaca bersamaan saat menampilkan portofolio, pendekatan embedding meminimalkan query round-trip ke database dan menjaga integritas data tanpa memerlukan relasi multi-tabel terpisah.

## 1. Contoh Dokumen MongoDB Lengkap

```javascript
{
  _id: ObjectId("671cc815e489620295e8025a"),
  student_id: 1,
  headline: "Fullstack Developer & Backend Specialist",
  bio: "Mahasiswa Teknik Informatika yang berfokus pada pengembangan backend scalable dengan Go, arsitektur REST API Fiber, dan manajemen dual-database PostgreSQL + MongoDB.",
  projects: [
    {
      _id: ObjectId("671cc815e489620295e8025b"),
      title: "SIAKAD Mini API",
      technologies: ["go", "mongodb", "postgresql", "fiber"],
      url: "https://github.com/example/siakad-mini",
      year: 2026
    },
    {
      _id: ObjectId("671cc815e489620295e8025c"),
      title: "Realtime Notification Engine",
      technologies: ["go", "redis", "docker"],
      url: "https://github.com/example/notif-engine",
      year: 2025
    }
  ],
  certificates: [
    {
      _id: ObjectId("671cc815e489620295e8025d"),
      name: "Certified Backend Specialist",
      issuer: "Go Academy",
      issued_at: ISODate("2026-01-15T00:00:00.000Z")
    }
  ],
  created_at: ISODate("2026-10-26T20:00:00.000Z"),
  updated_at: ISODate("2026-10-26T21:00:00.000Z")
}
```

## 2. Struktur Field Dokumen

| Field | Tipe | Keterangan |
|---|---|---|
| `_id` | `ObjectId` | Primary key dokumen di MongoDB |
| `student_id` | `int` | Relasi logis ke tabel `students.id` di PostgreSQL (Index Unik) |
| `headline` | `string` | Judul singkat profil portofolio (misal: "Fullstack Engineer") |
| `bio` | `string` | Deskripsi ringkas profil profesional |
| `projects` | `Array<Object>` | Daftar proyek hasil karya mahasiswa (maksimal 20 proyek) |
| `projects._id` | `ObjectId` | ID unik untuk tiap item proyek |
| `projects.title` | `string` | Judul / nama proyek |
| `projects.technologies` | `Array<string>` | Daftar teknologi yang digunakan (Multikey Index) |
| `projects.url` | `string` | Tautan repositori atau demo proyek |
| `projects.year` | `int` | Tahun pengerjaan proyek |
| `certificates` | `Array<Object>` | Daftar sertifikat kompetensi mahasiswa |
| `certificates._id` | `ObjectId` | ID unik tiap sertifikat |
| `certificates.name` | `string` | Nama sertifikat |
| `certificates.issuer` | `string` | Lembaga penerbit sertifikat |
| `certificates.issued_at` | `ISODate` | Tanggal terbit sertifikat |
| `created_at` | `ISODate` | Waktu pertama kali portofolio dibuat |
| `updated_at` | `ISODate` | Waktu terakhir pembaruan portofolio |

## 3. Strategi Indexing
1. **Unique Index**: `{ student_id: 1 }` (name: `student_portfolio_unique`)  
   Memastikan satu mahasiswa hanya memiliki satu dokumen portofolio.
2. **Multikey Index**: `{ "projects.technologies": 1 }` (name: `portfolio_tech_multikey`)  
   Mempercepat pencarian dan agregasi statistik teknologi terpopuler.
