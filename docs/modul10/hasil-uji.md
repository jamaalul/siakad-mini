# Hasil Uji Endpoint Ulasan (Modul 10)

Dokumentasi hasil pengujian endpoint ulasan kelas (PostgreSQL + MongoDB).

### Kasus 1 - Mahasiswa tidak terdaftar menulis ulasan (403)

**Perintah:**
```powershell
curl.exe -i -s -X POST "http://localhost:3000/api/v1/courses/1/reviews" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAzQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiNCIsImV4cCI6MTc5MzAzMDkxMywiaWF0IjoxNzkzMDI3MzEzfQ.IDe4BpTtzB2OqGiYD705zl5-2HxQ6uAdQzakyMGN2RU" -H "Content-Type: application/json" -d "@C:\Users\ajama\AppData\Local\Temp\m10\review1.json"
```
**Hasil:**
```http
HTTP/1.1 403 Forbidden
Date: Mon, 26 Oct 2026 15:08:33 GMT
Content-Type: application/json
Content-Length: 75
X-Request-Id: 8ae8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"success":false,"message":"hanya peserta kelas yang boleh menulis ulasan"}
```
### Kasus 2 - Mahasiswa terdaftar menulis ulasan (201)

**Perintah:**
```powershell
curl.exe -i -s -X POST "http://localhost:3000/api/v1/courses/1/reviews" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAxQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiMiIsImV4cCI6MTc5MzAzMDkxMiwiaWF0IjoxNzkzMDI3MzEyfQ.s8lYs9i_suF1LOcX2fKucVWCbPvlS97J0hda29GYepA" -H "Content-Type: application/json" -d "@C:\Users\ajama\AppData\Local\Temp\m10\review1.json"
```
**Hasil:**
```http
HTTP/1.1 201 Created
Date: Mon, 26 Oct 2026 15:08:33 GMT
Content-Type: application/json
Content-Length: 364
X-Request-Id: 8be8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none
Location: /api/v1/reviews/6adf6cf2e489620295e80258

{"data":{"id":"6adf6cf2e489620295e80258","course_id":1,"student_id":1,"student_name":"Ahmad Fauzi","rating":4,"comment":"Materi Go-nya jelas, tugasnya cukup banyak","tags":["tugas-banyak","dosenjelas"],"replies":[],"helpful_count":0,"created_at":"2026-10-26T15:08:34.34Z","updated_at":"2026-10-26T15:08:34.34Z"},"message":"ulasan berhasil disimpan","success":true}
```
### Kasus 3 - Ulasan kedua oleh mahasiswa yang sama pada kelas yang sama (409)

**Perintah:**
```powershell
curl.exe -i -s -X POST "http://localhost:3000/api/v1/courses/1/reviews" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAxQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiMiIsImV4cCI6MTc5MzAzMDkxMiwiaWF0IjoxNzkzMDI3MzEyfQ.s8lYs9i_suF1LOcX2fKucVWCbPvlS97J0hda29GYepA" -H "Content-Type: application/json" -d "@C:\Users\ajama\AppData\Local\Temp\m10\review1.json"
```
**Hasil:**
```http
HTTP/1.1 409 Conflict
Date: Mon, 26 Oct 2026 15:08:34 GMT
Content-Type: application/json
Content-Length: 93
X-Request-Id: 8ce8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"success":false,"message":"Anda sudah menulis ulasan untuk kelas ini, ubah ulasan yang ada"}
```
### Kasus 4 - Validasi payload tidak valid (422)

**Perintah:**
```powershell
curl.exe -i -s -X POST "http://localhost:3000/api/v1/courses/1/reviews" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAxQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiMiIsImV4cCI6MTc5MzAzMDkxMiwiaWF0IjoxNzkzMDI3MzEyfQ.s8lYs9i_suF1LOcX2fKucVWCbPvlS97J0hda29GYepA" -H "Content-Type: application/json" -d "@C:\Users\ajama\AppData\Local\Temp\m10\bad.json"
```
**Hasil:**
```http
HTTP/1.1 422 Unprocessable Entity
Date: Mon, 26 Oct 2026 15:08:34 GMT
Content-Type: application/json
Content-Length: 146
X-Request-Id: 8de8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"success":false,"message":"validasi gagal","errors":{"comment":["minimal 10 karakter"],"rating":["nilai maksimal 5"],"tags":["maksimal 5 item"]}}
```
### Kasus 5 - Kelas tidak ditemukan (404)

**Perintah:**
```powershell
curl.exe -i -s -X POST "http://localhost:3000/api/v1/courses/99999/reviews" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAxQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiMiIsImV4cCI6MTc5MzAzMDkxMiwiaWF0IjoxNzkzMDI3MzEyfQ.s8lYs9i_suF1LOcX2fKucVWCbPvlS97J0hda29GYepA" -H "Content-Type: application/json" -d "@C:\Users\ajama\AppData\Local\Temp\m10\review1.json"
```
**Hasil:**
```http
HTTP/1.1 404 Not Found
Date: Mon, 26 Oct 2026 15:08:34 GMT
Content-Type: application/json
Content-Length: 50
X-Request-Id: 8ee8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"success":false,"message":"data tidak ditemukan"}
```
### Kasus 6 - Format ObjectID tidak valid (400)

**Perintah:**
```powershell
curl.exe -i -s -X GET "http://localhost:3000/api/v1/reviews/bukan-objectid" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAxQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiMiIsImV4cCI6MTc5MzAzMDkxMiwiaWF0IjoxNzkzMDI3MzEyfQ.s8lYs9i_suF1LOcX2fKucVWCbPvlS97J0hda29GYepA"
```
**Hasil:**
```http
HTTP/1.1 400 Bad Request
Date: Mon, 26 Oct 2026 15:08:34 GMT
Content-Type: application/json
Content-Length: 51
X-Request-Id: 8fe8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"success":false,"message":"id ulasan tidak valid"}
```
### Kasus 7 - Mahasiswa lain mengubah ulasan orang lain (403)

**Perintah:**
```powershell
curl.exe -i -s -X PATCH "http://localhost:3000/api/v1/reviews/6adf6cf2e489620295e80258" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAyQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiMyIsImV4cCI6MTc5MzAzMDkxMiwiaWF0IjoxNzkzMDI3MzEyfQ.C3Cd18onoIPem_oOqrpq8AhHGPryCNW-3KomQyfNK_M" -H "Content-Type: application/json" -d "@C:\Users\ajama\AppData\Local\Temp\m10\patch.json"
```
**Hasil:**
```http
HTTP/1.1 403 Forbidden
Date: Mon, 26 Oct 2026 15:08:35 GMT
Content-Type: application/json
Content-Length: 75
X-Request-Id: 90e8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"success":false,"message":"hanya pemilik ulasan yang boleh melakukan ini"}
```
### Kasus 8 - Pemilik mengubah ulasan sendiri (200)

**Perintah:**
```powershell
curl.exe -i -s -X PATCH "http://localhost:3000/api/v1/reviews/6adf6cf2e489620295e80258" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAxQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiMiIsImV4cCI6MTc5MzAzMDkxMiwiaWF0IjoxNzkzMDI3MzEyfQ.s8lYs9i_suF1LOcX2fKucVWCbPvlS97J0hda29GYepA" -H "Content-Type: application/json" -d "@C:\Users\ajama\AppData\Local\Temp\m10\patch.json"
```
**Hasil:**
```http
HTTP/1.1 200 OK
Date: Mon, 26 Oct 2026 15:08:35 GMT
Content-Type: application/json
Content-Length: 362
X-Request-Id: 91e8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"data":{"id":"6adf6cf2e489620295e80258","course_id":1,"student_id":1,"student_name":"Ahmad Fauzi","rating":5,"comment":"Materi Go-nya jelas, tugasnya cukup banyak","tags":["tugas-banyak","dosenjelas"],"replies":[],"helpful_count":0,"created_at":"2026-10-26T15:08:34.34Z","updated_at":"2026-10-26T15:08:35.89Z"},"message":"ulasan berhasil diubah","success":true}
```
### Kasus 9 - Menilai ulasan sendiri sebagai membantu (409)

**Perintah:**
```powershell
curl.exe -i -s -X POST "http://localhost:3000/api/v1/reviews/6adf6cf2e489620295e80258/helpful" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAxQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiMiIsImV4cCI6MTc5MzAzMDkxMiwiaWF0IjoxNzkzMDI3MzEyfQ.s8lYs9i_suF1LOcX2fKucVWCbPvlS97J0hda29GYepA"
```
**Hasil:**
```http
HTTP/1.1 409 Conflict
Date: Mon, 26 Oct 2026 15:08:35 GMT
Content-Type: application/json
Content-Length: 64
X-Request-Id: 92e8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"success":false,"message":"tidak dapat menilai ulasan sendiri"}
```
### Kasus 10 - Mahasiswa lain menilai membantu pertama (200) lalu mengulang (409)

**Perintah:**
```powershell
curl.exe -i -s -X POST "http://localhost:3000/api/v1/reviews/6adf6cf2e489620295e80258/helpful" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAyQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiMyIsImV4cCI6MTc5MzAzMDkxMiwiaWF0IjoxNzkzMDI3MzEyfQ.C3Cd18onoIPem_oOqrpq8AhHGPryCNW-3KomQyfNK_M" ; curl.exe -i -s -X POST "http://localhost:3000/api/v1/reviews/6adf6cf2e489620295e80258/helpful" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAyQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiMyIsImV4cCI6MTc5MzAzMDkxMiwiaWF0IjoxNzkzMDI3MzEyfQ.C3Cd18onoIPem_oOqrpq8AhHGPryCNW-3KomQyfNK_M"
```
**Hasil:**
```http
HTTP/1.1 200 OK
Date: Mon, 26 Oct 2026 15:08:35 GMT
Content-Type: application/json
Content-Length: 370
X-Request-Id: 93e8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"data":{"id":"6adf6cf2e489620295e80258","course_id":1,"student_id":1,"student_name":"Ahmad Fauzi","rating":5,"comment":"Materi Go-nya jelas, tugasnya cukup banyak","tags":["tugas-banyak","dosenjelas"],"replies":[],"helpful_count":1,"created_at":"2026-10-26T15:08:34.34Z","updated_at":"2026-10-26T15:08:35.89Z"},"message":"terima kasih atas penilaiannya","success":true}
HTTP/1.1 409 Conflict
Date: Mon, 26 Oct 2026 15:08:36 GMT
Content-Type: application/json
Content-Length: 59
X-Request-Id: 94e8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"success":false,"message":"Anda sudah menilai ulasan ini"}
```
### Kasus 11 - Menambahkan balasan (201)

**Perintah:**
```powershell
curl.exe -i -s -X POST "http://localhost:3000/api/v1/reviews/6adf6cf2e489620295e80258/replies" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAyQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiMyIsImV4cCI6MTc5MzAzMDkxMiwiaWF0IjoxNzkzMDI3MzEyfQ.C3Cd18onoIPem_oOqrpq8AhHGPryCNW-3KomQyfNK_M" -H "Content-Type: application/json" -d "@C:\Users\ajama\AppData\Local\Temp\m10\reply.json"
```
**Hasil:**
```http
HTTP/1.1 201 Created
Date: Mon, 26 Oct 2026 15:08:36 GMT
Content-Type: application/json
Content-Length: 526
X-Request-Id: 95e8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"data":{"id":"6adf6cf2e489620295e80258","course_id":1,"student_id":1,"student_name":"Ahmad Fauzi","rating":5,"comment":"Materi Go-nya jelas, tugasnya cukup banyak","tags":["tugas-banyak","dosenjelas"],"replies":[{"id":"6adf6cf4e489620295e8025a","student_id":2,"student_name":"Budi Santoso","body":"Setuju, tugasnya memang banyak","created_at":"2026-10-26T15:08:36.902Z"}],"helpful_count":1,"created_at":"2026-10-26T15:08:34.34Z","updated_at":"2026-10-26T15:08:35.89Z"},"message":"balasan berhasil ditambahkan","success":true}
```
### Kasus 12 - Mengambil ringkasan ulasan kelas (200)

**Perintah:**
```powershell
curl.exe -i -s -X GET "http://localhost:3000/api/v1/courses/1/reviews/summary" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAxQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiMiIsImV4cCI6MTc5MzAzMDkxMiwiaWF0IjoxNzkzMDI3MzEyfQ.s8lYs9i_suF1LOcX2fKucVWCbPvlS97J0hda29GYepA"
```
**Hasil:**
```http
HTTP/1.1 200 OK
Date: Mon, 26 Oct 2026 15:08:36 GMT
Content-Type: application/json
Content-Length: 238
X-Request-Id: 96e8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"data":{"course_id":1,"count":1,"average_rating":5,"distribution":{"1":0,"2":0,"3":0,"4":0,"5":1},"top_tags":[{"tag":"dosenjelas","total":1},{"tag":"tugas-banyak","total":1}]},"message":"ringkasan ulasan berhasil diambil","success":true}
```
### Kasus 13 - Filter ulasan berdasarkan tag (200)

**Perintah:**
```powershell
curl.exe -i -s -X GET "http://localhost:3000/api/v1/courses/1/reviews?tag=tugas-banyak" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAxQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiMiIsImV4cCI6MTc5MzAzMDkxMiwiaWF0IjoxNzkzMDI3MzEyfQ.s8lYs9i_suF1LOcX2fKucVWCbPvlS97J0hda29GYepA"
```
**Hasil:**
```http
HTTP/1.1 200 OK
Date: Mon, 26 Oct 2026 15:08:37 GMT
Content-Type: application/json
Content-Length: 584
X-Request-Id: 97e8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"data":[{"id":"6adf6cf2e489620295e80258","course_id":1,"student_id":1,"student_name":"Ahmad Fauzi","rating":5,"comment":"Materi Go-nya jelas, tugasnya cukup banyak","tags":["tugas-banyak","dosenjelas"],"replies":[{"id":"6adf6cf4e489620295e8025a","student_id":2,"student_name":"Budi Santoso","body":"Setuju, tugasnya memang banyak","created_at":"2026-10-26T15:08:36.902Z"}],"helpful_count":1,"created_at":"2026-10-26T15:08:34.34Z","updated_at":"2026-10-26T15:08:35.89Z"}],"message":"daftar ulasan berhasil diambil","meta":{"has_more":false,"limit":10,"next_before":""},"success":true}
```
### Kasus 14 - Pencegahan injeksi query string operator (200 dengan 0 hasil)

**Perintah:**
```powershell
curl.exe -i -s -X GET "http://localhost:3000/api/v1/courses/1/reviews?tag={%22$ne%22:%22%22}" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAxQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiMiIsImV4cCI6MTc5MzAzMDkxMiwiaWF0IjoxNzkzMDI3MzEyfQ.s8lYs9i_suF1LOcX2fKucVWCbPvlS97J0hda29GYepA"
```
**Hasil:**
```http
HTTP/1.1 200 OK
Date: Mon, 26 Oct 2026 15:08:37 GMT
Content-Type: application/json
Content-Length: 123
X-Request-Id: 98e8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"data":[],"message":"daftar ulasan berhasil diambil","meta":{"has_more":false,"limit":10,"next_before":""},"success":true}
```
### Kasus 15 - Pencegahan injeksi operator NoSQL di body (400)

**Perintah:**
```powershell
curl.exe -i -s -X POST "http://localhost:3000/api/v1/courses/1/reviews" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAyQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiMyIsImV4cCI6MTc5MzAzMDkxMiwiaWF0IjoxNzkzMDI3MzEyfQ.C3Cd18onoIPem_oOqrpq8AhHGPryCNW-3KomQyfNK_M" -H "Content-Type: application/json" -d "@C:\Users\ajama\AppData\Local\Temp\m10\inject.json"
```
**Hasil:**
```http
HTTP/1.1 400 Bad Request
Date: Mon, 26 Oct 2026 15:08:37 GMT
Content-Type: application/json
Content-Length: 63
X-Request-Id: 99e8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"success":false,"message":"body harus berupa JSON yang valid"}
```
### Kasus 16 - Bukan pemilik mencoba menghapus ulasan (403)

**Perintah:**
```powershell
curl.exe -i -s -X DELETE "http://localhost:3000/api/v1/reviews/6adf6cf2e489620295e80258" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1haGFzaXN3YTAzQHNpYWthZC50ZXN0Iiwicm9sZSI6Im1haGFzaXN3YSIsImlzcyI6InNpYWthZC1taW5pIiwic3ViIjoiNCIsImV4cCI6MTc5MzAzMDkxMywiaWF0IjoxNzkzMDI3MzEzfQ.IDe4BpTtzB2OqGiYD705zl5-2HxQ6uAdQzakyMGN2RU"
```
**Hasil:**
```http
HTTP/1.1 403 Forbidden
Date: Mon, 26 Oct 2026 15:08:37 GMT
Content-Type: application/json
Content-Length: 75
X-Request-Id: 9ae8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"success":false,"message":"hanya pemilik ulasan yang boleh melakukan ini"}
```
### Kasus 17 - Admin menghapus ulasan siapa pun (204)

**Perintah:**
```powershell
curl.exe -i -s -X DELETE "http://localhost:3000/api/v1/reviews/6adf6cf2e489620295e80258" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFkbWluQHNpYWthZC50ZXN0Iiwicm9sZSI6ImFkbWluIiwiaXNzIjoic2lha2FkLW1pbmkiLCJzdWIiOiIxIiwiZXhwIjoxNzkzMDMwOTEzLCJpYXQiOjE3OTMwMjczMTN9.ulL2SMON-HyG9bX9wzeot_nRc3T8MF5rSLuXOYG48Wo"
```
**Hasil:**
```http
HTTP/1.1 204 No Content
Date: Mon, 26 Oct 2026 15:08:38 GMT
X-Request-Id: 9be8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

```
### Kasus 18 - Menghapus kembali ulasan yang sudah dihapus (404)

**Perintah:**
```powershell
curl.exe -i -s -X DELETE "http://localhost:3000/api/v1/reviews/6adf6cf2e489620295e80258" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFkbWluQHNpYWthZC50ZXN0Iiwicm9sZSI6ImFkbWluIiwiaXNzIjoic2lha2FkLW1pbmkiLCJzdWIiOiIxIiwiZXhwIjoxNzkzMDMwOTEzLCJpYXQiOjE3OTMwMjczMTN9.ulL2SMON-HyG9bX9wzeot_nRc3T8MF5rSLuXOYG48Wo"
```
**Hasil:**
```http
HTTP/1.1 404 Not Found
Date: Mon, 26 Oct 2026 15:08:38 GMT
Content-Type: application/json
Content-Length: 50
X-Request-Id: 9ce8a92b-7407-42fb-b6c0-d3e9e0c66a6d
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
Origin-Agent-Cluster: ?1
Referrer-Policy: no-referrer
X-Dns-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none

{"success":false,"message":"data tidak ditemukan"}
```
## Uji Serentak

### Uji 1 - 20 Penilaian Membantu Serentak oleh Satu Mahasiswa

Mahasiswa 2 mengirimkan 20 permintaan serentak untuk menandai ulasan sebagai membantu (`POST /reviews/:id/helpful`).
Hanya tepat 1 permintaan yang berhasil (200), sedangkan 19 permintaan lainnya ditolak dengan status 409 Conflict karena operasi atomik `$addToSet` dan filter `helpful_by: {$ne: studentID}` mencegah duplikasi voting secara atomik.

**Hasil Pengelompokan Kode Status:**
```text
Count Name
----- ----
    1 200
   19 409
```

### Uji 2 - Batas 50 Balasan dengan 10 Permintaan Serentak

Dokumen ulasan sebelumnya telah diisi 48 balasan. Kemudian dikirim 10 permintaan balasan serentak (`POST /reviews/:id/replies`).
Tepat 2 permintaan berhasil (201) mengisi slot ke-49 dan ke-50, sedangkan 8 permintaan lainnya ditolak (409 Conflict) karena pemeriksaan atomik `replies.49: {$exists: false}` berhasil menjaga batas maksimal 50 balasan.

**Hasil Pengelompokan Kode Status:**
```text
Count Name
----- ----
    2 201
    8 409
```
