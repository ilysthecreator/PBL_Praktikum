package main

import "fmt"

func main() {

	var nama string = "Muhammad Fadhil Ilyas"
	var umur int = 21 
	var ipk float64 = 3.85
	var statusMahasiswa bool = true

	makananFavorit := []string{"Ayam Geprek", "Soto Ayam", "Dimsum"}

	fmt.Println("Data Diri")
	fmt.Println("Nama:", nama)
	fmt.Println("Umur:", umur)
	fmt.Println("IPK:", ipk)
	fmt.Println("Aktif kuliah?", statusMahasiswa)
	fmt.Println("Suka makan:", makananFavorit)
	fmt.Println()

	nilaiMahasiswa := make(map[string]int)

	nilaiMahasiswa["Muhammad Fadhil Ilyas"] = 95
	nilaiMahasiswa["Ody Dzakwan berwin"] = 80
	nilaiMahasiswa["Muhammad Reyhan Ammar"] = 85

	fmt.Println("Cek Data Map")
	nilai, ada := nilaiMahasiswa["Muhammad Fadhil Ilyas"]
	if ada == true {
		fmt.Println("Data ketemu! Nilai Fadhil adalah:", nilai)
	} else {
		fmt.Println("Data Fadhil nggak ketemu.")
	}

	_, adaAlfin := nilaiMahasiswa["Alfin Dhanur Rianto"]
	if adaAlfin == true {
		fmt.Println("Alfin Dhanur Rianto ada di data")
	} else {
		fmt.Println("Data Alfin Dhanur Rianto nggak ditemukan.")
	}

	delete(nilaiMahasiswa, "Ody Dzakwan berwin")
	fmt.Println("\nInfo: Data Ody Dzakwan berwin baru saja dihapus.")

	fmt.Println("\nDaftar Nilai Akhir")
	for kunci, isi := range nilaiMahasiswa {
		fmt.Println("- Nama:", kunci, "| Nilai:", isi)
	}
}