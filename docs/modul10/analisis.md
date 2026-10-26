# Analisis Tugas Mandiri Modul 10 — Portofolio Mahasiswa

## 1. Mengapa `projects` Disarangkan (Embedded), dan Satu Perubahan Kebutuhan yang Membalik Keputusan Itu

Pada rancangan ini, `projects` disimpan sebagai array di dalam dokumen `student_portfolios` (embedding) karena tiga alasan utama. Pertama, proyek bersifat **eksklusif milik satu mahasiswa** — tidak ada mahasiswa lain yang mereferensikan proyek yang sama, sehingga tidak ada kebutuhan relasi many-to-many. Kedua, proyek **selalu dibaca bersamaan** dengan data portofolio; saat menampilkan profil seorang mahasiswa, seluruh proyeknya ikut ditampilkan sehingga embedding menghilangkan query tambahan ke collection lain. Ketiga, jumlah proyek **dibatasi 20 item** sehingga dokumen tidak akan tumbuh tak terbatas dan tidak akan mendekati batas 16 MB per dokumen MongoDB.

Keputusan ini berbalik jika muncul kebutuhan **pencarian lintas mahasiswa berdasarkan proyek tertentu**, misalnya: "Tampilkan semua mahasiswa yang mengerjakan proyek bertema machine learning" atau "Cari proyek dengan judul mengandung kata kunci X." Dengan embedding, setiap query semacam itu harus memindai semua dokumen `student_portfolios` dan membuka array `projects` di tiap dokumen — operasi yang mahal. Jika kebutuhan ini muncul, pendekatan yang lebih tepat adalah memisahkan `projects` ke collection sendiri (`student_projects`) dengan field `student_id` sebagai foreign key, lalu membuat index teks pada field `title` dan `technologies`. Dengan demikian query pencarian dapat menggunakan index langsung tanpa full collection scan.

---

## 2. Dampak Menghapus Mahasiswa di PostgreSQL terhadap Portofolio di MongoDB

Karena antara PostgreSQL dan MongoDB **tidak ada foreign key lintas database**, menghapus baris di tabel `students` pada PostgreSQL tidak otomatis menghapus dokumen di collection `student_portfolios` pada MongoDB. Akibatnya, dokumen portofolio tetap tersimpan di MongoDB meski mahasiswanya sudah tidak ada, membentuk **orphan document** — data yang tidak lagi memiliki pemilik yang valid.

Dua cara menanganinya beserta kelemahannya:

**Cara 1 — Soft delete + cron job pembersihan.** Mahasiswa tidak benar-benar dihapus dari PostgreSQL, melainkan ditandai `deleted_at`. Secara berkala, sebuah job memeriksa mahasiswa yang sudah di-soft-delete dan menghapus dokumen MongoDB-nya.
*Kelemahan:* Membutuhkan mekanisme penjadwalan tambahan; ada jeda waktu antara mahasiswa dihapus dan dokumen portofolionya ikut dibersihkan — selama jeda itu data orphan masih ada.

**Cara 2 — Hapus portofolio secara eksplisit di layer service sebelum menghapus mahasiswa.** Endpoint `DELETE /students/:id` di service memanggil `portfolioRepo.DeleteByStudent` terlebih dahulu sebelum menghapus baris di PostgreSQL, semuanya dalam satu transaksi logis (meskipun bukan transaksi ACID lintas database).
*Kelemahan:* Jika langkah penghapusan portofolio di MongoDB berhasil tetapi penghapusan mahasiswa di PostgreSQL gagal, terjadi inkonsistensi terbalik — mahasiswanya masih ada tetapi portofolionya sudah hilang. Dibutuhkan strategi rollback atau idempotency yang cermat.

---

## 3. Endpoint yang Tetap Berjalan saat MongoDB Mati

Saat service MongoDB dimatikan (`Stop-Service MongoDB`), endpoint yang **tetap berjalan** adalah semua endpoint yang hanya bergantung pada **PostgreSQL**:

- `POST /api/v1/auth/login` — autentikasi membaca tabel `users` di PostgreSQL ✅
- `GET /api/v1/auth/me` — membaca data user dan student dari PostgreSQL ✅
- `GET /api/v1/students` — membaca tabel `students` di PostgreSQL ✅
- `GET /api/v1/courses` — membaca tabel `courses` di PostgreSQL ✅
- `POST /api/v1/enrollments` — membaca dan menulis tabel `enrollments` di PostgreSQL ✅

Sementara itu, **endpoint yang langsung gagal** saat MongoDB mati:

- `PUT /students/:id/portfolio` → 500 (timeout koneksi MongoDB) ❌
- `GET /students/:id/portfolio` → 500 ❌
- `POST /students/:id/portfolio/projects` → 500 ❌
- `GET /api/v1/reviews/...` → 500 ❌
- `GET /api/v1/portfolios/stats/technologies` → 500 ❌
- `GET /health` → **503 Service Unavailable** (health check mendeteksi MongoDB tidak dapat dihubungi) ❌

Hal ini menunjukkan arsitektur *dual-database* memiliki implikasi ketersediaan: fitur ulasan dan portofolio sepenuhnya bergantung pada MongoDB, sedangkan fitur akademik inti (autentikasi, data mahasiswa, KRS) tetap berfungsi selama PostgreSQL hidup. Untuk produksi, perlu dipertimbangkan *circuit breaker* agar kegagalan MongoDB tidak menyebabkan seluruh aplikasi merespons 500.
