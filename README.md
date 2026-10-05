# Peminjaman Buku API

Ini adalah project REST API untuk **peminjaman buku**, dibuat pakai bahasa Go
Dengan API ini, user bisa daftar, login, melihat buku, dan meminjam buku. Admin bisa mengelola buku, user, dan pengembalian

---

## Fitur

- Daftar akun dan login (pakai JWT)
- Refresh token dan logout
- Lihat, tambah, ubah, dan hapus buku
- Pinjam buku dan kembalikan buku
- Hak akses berdasarkan role (`admin` dan `user`)
- Daftar pinjaman bisa diunduh dalam bentuk CSV
- Batas login supaya tidak bisa ditebak-tebak (rate limit)
- Log tersimpan di file `logs/app.log`

## Teknologi yang dipakai

| Apa | Dipakai untuk |
|---|---|
| Go | Bahasa pemrograman |
| Fiber v2 | Framework web |
| PostgreSQL | Database |
| pgx | Menghubungkan Go ke PostgreSQL |
| JWT (golang-jwt) | Token login |
| bcrypt | Mengacak (hash) password |
| validator | Mengecek isi request |
| slog + lumberjack | Menulis log |

## Struktur folder

```
PeminjamanBuku/
├── main.go          -> titik awal program
├── config/          -> pengaturan (env, logger, app)
├── database/        -> koneksi ke PostgreSQL
├── route/           -> daftar URL (endpoint)
├── middleware/      -> pengecekan login, hak akses, rate limit
├── app/
│   ├── model/       -> bentuk data (struct)
│   ├── repository/  -> perintah ke database
│   └── service/     -> logika dan handler
├── helper/          -> fungsi bantu (JWT, response, validasi, dll)
├── migrations/      -> file SQL untuk membuat tabel
├── .env.example     -> contoh isi file .env
└── logs/            -> file log (dibuat otomatis)
```

---

## Cara menjalankan

### 1. Yang harus sudah terpasang

- [Go](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/) (sudah termasuk `psql`)

### 2. Download project

```bash
git clone https://github.com/Sevynix/PeminjamanBuku.git
cd PeminjamanBuku
```

### 3. Buat database

Masuk ke PostgreSQL, lalu buat database kosong:

```sql
CREATE DATABASE peminjaman_buku;
```

### 4. Buat tabel (jalankan file migrations)

Jalankan **berurutan** dari 001 sampai 005:

```bash
psql -h localhost -U postgres -d peminjaman_buku -f migrations/001_create_books.sql
psql -h localhost -U postgres -d peminjaman_buku -f migrations/002_auth.sql
psql -h localhost -U postgres -d peminjaman_buku -f migrations/003_rbac.sql
psql -h localhost -U postgres -d peminjaman_buku -f migrations/004_create_loans.sql
psql -h localhost -U postgres -d peminjaman_buku -f migrations/005_loans_cursor_index.sql
```

Urutannya penting. Kalau terbalik, bisa error.

### 5. Buat file `.env`

Salin `.env.example` menjadi `.env`, lalu isi:

```env
APP_NAME=Peminjaman Buku API
APP_PORT=3000
LOG_LEVEL=info

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=isi_password_postgres_kamu
DB_NAME=peminjaman_buku
DB_SSLMODE=disable
DB_MAX_CONNS=10

JWT_SECRET=isi_dengan_teks_acak_minimal_32_karakter
JWT_ISSUER=peminjaman-buku
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_DAYS=7

ALLOWED_ORIGINS=http://localhost:5173
```

Hal yang perlu diperhatikan:

- `JWT_SECRET` **minimal 32 karakter**. Kalau kurang, server menolak jalan
- `.env` jangan di-upload ke GitHub, karena berisi password. File ini sudah ada di `.gitignore`

### 6. Jalankan server

```bash
go mod tidy
go run main.go
```

Kalau berhasil, server jalan di `http://localhost:3000`.
Cek dengan membuka `http://localhost:3000/api/v1/health` di browser

---

## Cara membuat akun admin

Saat register, role selalu `user`. Ini disengaja supaya orang tidak bisa mendaftar sendiri sebagai admin.

Untuk membuat admin pertama, ubah role-nya langsung di database:

```sql
UPDATE users SET role = 'admin' WHERE username = 'nama_user_kamu';
```

Setelah itu **login ulang**, karena role ikut tersimpan di token

---

## Daftar endpoint

Semua URL diawali `/api/v1`. Kolom **Akses**:

- **Publik** = tanpa login
- **Login** = harus kirim token
- **Admin** = harus punya permission tertentu
- **Pemilik/Admin** = boleh kalau datanya milik sendiri, atau kalau admin

### Auth

| Method | URL | Akses | Fungsi |
|---|---|---|---|
| POST | `/auth/register` | Publik | Daftar akun baru |
| POST | `/auth/login` | Publik | Login (maksimal 5x per menit) |
| POST | `/auth/refresh` | Publik | Tukar refresh token dengan token baru |
| POST | `/auth/logout` | Publik | Logout (refresh token dibatalkan) |
| GET | `/auth/me` | Login | Lihat profil dan daftar permission |

### Buku

| Method | URL | Akses | Fungsi |
|---|---|---|---|
| GET | `/books` | Login | Daftar buku |
| GET | `/books/:id` | Login | Detail satu buku |
| POST | `/books` | Admin | Tambah buku |
| PUT | `/books/:id` | Admin | Ganti seluruh data buku |
| PATCH | `/books/:id` | Admin | Ubah sebagian data buku |
| DELETE | `/books/:id` | Admin | Hapus buku |

### User

| Method | URL | Akses | Fungsi |
|---|---|---|---|
| GET | `/users` | Admin | Daftar semua user |
| GET | `/users/:id` | Pemilik/Admin | Detail user |
| PATCH | `/users/:id` | Pemilik/Admin | Ubah username atau email |
| PATCH | `/users/:id/role` | Admin | Ubah role user |
| DELETE | `/users/:id` | Admin | Hapus user |

### Pinjaman

| Method | URL | Akses | Fungsi |
|---|---|---|---|
| POST | `/loans` | Login | Pinjam buku |
| GET | `/loans/me` | Login | Daftar pinjaman milik sendiri |
| GET | `/loans` | Admin | Daftar semua pinjaman |
| GET | `/loans/:id` | Pemilik/Admin | Detail satu pinjaman |
| PATCH | `/loans/:id/return` | Admin | Proses pengembalian buku |

### Parameter tambahan di URL

- **Buku dan user**: `page`, `limit`, `search`, `sort`, `order` (`asc` atau `desc`). Khusus buku ada `available=true`, khusus user ada `role=admin`
- **Pinjaman**: `limit`, `cursor` (untuk halaman berikutnya), dan `status` (`active`, `returned`, atau `overdue`)
- **Unduh CSV**: kirim header `Accept: text/csv` ke `/loans` atau `/loans/me`

---

## Aturan peminjaman

- Satu user maksimal **3 pinjaman aktif** sekaligus
- Lama pinjam **7 hari**. Lewat dari itu statusnya `overdue`
- Buku yang stoknya 0 tidak bisa dipinjam
- Buku yang sama tidak bisa dipinjam dua kali selagi belum dikembalikan
- Saat buku dipinjam stok berkurang 1, saat dikembalikan stok bertambah 1
- Buku atau user yang punya riwayat pinjaman tidak bisa dihapus (hasilnya `409`)
- Admin tidak bisa menghapus akunnya sendiri

## Aturan akun

- Username: 3 sampai 30 karakter, hanya huruf, angka, titik, dan garis bawah. Tidak peduli huruf besar atau kecil (`Budi` dan `budi` dianggap sama)
- Password: minimal 8 karakter, harus ada huruf **dan** angka, dan tidak boleh password yang terlalu umum seperti `password123`

---

## Contoh pemakaian

Contoh ini memakai `curl`. Di Windows PowerShell, tulis `curl.exe` (bukan `curl`)

**Register**

```bash
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"budi","email":"budi@example.com","password":"Rahasia2026x"}'
```

**Login** (ambil `access_token` dari hasilnya)

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"budi","password":"Rahasia2026x"}'
```

**Pinjam buku** (ganti `TOKEN_KAMU` dengan access token)

```bash
curl -X POST http://localhost:3000/api/v1/loans \
  -H "Authorization: Bearer TOKEN_KAMU" \
  -H "Content-Type: application/json" \
  -d '{"book_id":1}'
```

### Bentuk response

Kalau berhasil:

```json
{
  "success": true,
  "message": "peminjaman berhasil",
  "data": { }
}
```

Kalau gagal:

```json
{
  "success": false,
  "code": "NOT_FOUND",
  "message": "buku tidak ditemukan",
  "request_id": "..."
}
```

### Arti kode status

| Kode | Artinya |
|---|---|
| 200 / 201 | Berhasil / berhasil dibuat |
| 400 | Request salah (misalnya JSON rusak) |
| 401 | Belum login atau token tidak valid |
| 403 | Tidak punya hak akses |
| 404 | Data tidak ditemukan |
| 406 | Format yang diminta tidak tersedia |
| 409 | Bentrok (data sudah ada, stok habis, atau masih direferensikan) |
| 415 | Content-Type bukan `application/json` |
| 422 | Isi data tidak lolos validasi |
| 429 | Terlalu banyak percobaan login |
| 500 / 503 | Masalah di server atau database |

---

## Pengujian

Pengujian dilakukan manual memakai `curl`. Hal yang sudah dicoba:

- Register, login, refresh token (token lama tidak bisa dipakai lagi), logout
- CRUD buku oleh admin, dan user biasa ditolak (403)
- Pinjam dan kembalikan buku, termasuk stok bertambah dan berkurang
- Batas 3 pinjaman aktif, stok habis, dan pinjam buku yang sama dua kali
- Hapus buku atau user yang punya pinjaman (409)
- Token yang diubah satu huruf ditolak (401)
- Pagination dengan cursor, dan unduh CSV
- Akses `Accept: application/xml` ditolak (406)
- Register dengan username kembar beda huruf besar/kecil (409) dan password umum (422)
- Register dengan `"role":"admin"` tetap menjadi `user`
- Batas login 5x per menit (429)
- Isi `logs/app.log`
- PostgreSQL dimatikan: `/health` menjawab 503 dan endpoint lain menjawab pesan umum tanpa detail database

## Catatan keamanan

- Password disimpan dalam bentuk hash (bcrypt), bukan teks asli
- Pesan login salah sama untuk "password salah" dan "username tidak ada", supaya username tidak bisa ditebak
- Refresh token disimpan dalam bentuk hash, dan hanya bisa dipakai satu kali
- Jangan upload `.env` atau folder `logs/` ke GitHub