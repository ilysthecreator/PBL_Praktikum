# Dokumentasi API - Students

REST API untuk manajemen data mahasiswa, dibuat menggunakan Go dan Fiber.

## Kontrak API

| Metode | Endpoint | Parameter | Contoh Body Permintaan | Status | Contoh Respons |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **GET** | `/api/v1/students` | Query: `page`, `limit`, `search`, `sort`, `order`, `is_active` | *(Kosong)* | `200 OK` | `{"success":true,"message":"...","data":[...],"meta":{"page":1,"limit":10,"total":5,"total_pages":1}}` |
| **GET** | `/api/v1/students/:id` | Path: `id` (int) | *(Kosong)* | `200 OK`<br>`400 Bad Request`<br>`404 Not Found` | `{"success":true,"message":"...","data":{"id":1,"nim":"123","name":"Budi","grade":85,"is_active":true}}` |
| **POST** | `/api/v1/students` | *(Kosong)* | `{"nim":"12345", "name":"Andi", "grade":90}` | `201 Created`<br>`400 Bad Request`<br>`409 Conflict`<br>`415 Unsupported Media Type`<br>`422 Unprocessable Entity` | `{"success":true,"message":"mahasiswa berhasil ditambahkan","data":{...}}` *(disertai header Location)* |
| **PUT** | `/api/v1/students/:id` | Path: `id` (int) | `{"nim":"12345", "name":"Andi Baru", "grade":95, "is_active":false}` | `200 OK`<br>`400 Bad Request`<br>`404 Not Found`<br>`409 Conflict`<br>`422 Unprocessable Entity` | `{"success":true,"message":"data mahasiswa berhasil diganti seluruhnya","data":{...}}` |
| **PATCH** | `/api/v1/students/:id` | Path: `id` (int) | `{"grade": 100}` | `200 OK`<br>`400 Bad Request`<br>`404 Not Found`<br>`409 Conflict`<br>`422 Unprocessable Entity` | `{"success":true,"message":"data mahasiswa berhasil diperbarui sebagian","data":{...}}` |
| **DELETE** | `/api/v1/students/:id` | Path: `id` (int) | *(Kosong)* | `204 No Content`<br>`400 Bad Request`<br>`404 Not Found` | *(Tidak ada body)* |
