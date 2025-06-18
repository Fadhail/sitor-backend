# SITOR Backend

Backend aplikasi SITOR menggunakan Golang, Fiber, dan MongoDB.

## Fitur
- Register user
- Login user

## Struktur data user
- name: string
- email: string (unik)
- password: string (hash)
- role: string (anggota/ketua)

## Endpoint
- POST `/api/register`
- POST `/api/login`

## Menjalankan
```
go run main.go
```
