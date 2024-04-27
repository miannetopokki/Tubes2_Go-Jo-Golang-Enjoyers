# Tugas Besar 2 IF2211 Strategi Algoritma Semester II tahun 2023/2024
# Pemanfaatan Algoritma IDS dan BFS dalam Permainan WikiRace

### Dibuat oleh:
| Nama | NIM |
| -------- | --------- |
| Hugo Sabam Augusto | 13522129 |
| Nicholas Reymond Sihite | 13522144 |
| Muhammad Rasheed Qais Tandjung | 13522158 |

## Deskripsi Program
Program ini adalah program berbasis web yang dibuat untuk mencari rute antara satu halaman wikipedia (source) ke halaman wikipedia lainnya (destination)
dengan menggunakan algoritma BFS (Breadth First Search) dan IDS (Iterative Deepening Search). Bahasa pemrograman yang digunakan dalam implementasi
algoritma dan back-end website-nya adalah Go dan bahasa yang digunakan pada front-end website-nya adalah HTML dan CSS.
Program akan menampilkan rute, waktu pencarian, derajat pencarian, dan banyak halaman yang dikunjungi.

## Requirement Program
Tidak ada kebutuhan khusus selain instalasi compiler Go

## Cara Menggunakan Program
1. Clone repository dengan cara membuka terminal lalu masukkan perintah berikut
   ```sh
   git clone https://github.com/miannetopokki/Tubes2_Go-Jo.git
   ```
2. Buka folder "src" lalu buka terminal (jika pada folder src tidak ada cache.txt, buat terlebih dahulu)
3. Jalankan web dengan perintah
    ```sh
    go run main.go IDSR.go BFS.go BFSCache.go queueLinked.go
    ```
4. Terminal akan menampilkan alamat tempat program dijalankan. Pengguna dapat melakukan ctrl + click pada teks di terminal atau membuka secara manual pada browser.
    ```sh
    localhost:8080
    ```
5. Akan ada dua masukan yang akan diterima program, yaitu halaman wikipedia awal dan tujuan.
6. Ada 2 tombol untuk memproses masukan, yaitu BFS dan IDS. Algoritma yang akan dijalankan bergantung pada tombol yang dipilih. 
7. Program akan menampilkan hasil pencarian, waktu pencarian, derajat pencarian, dan banyak halaman yang dikunjungi.

## Credits

Dalam pengembangan aplikasi ini, kami mengadopsi penggunaan teknologi caching untuk algoritma IDS dari library [bigcache](https://github.com/allegro/bigcache).
