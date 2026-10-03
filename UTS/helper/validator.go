package helper

import (
	"regexp"
	"strconv"
	"time"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	nimRegex      = regexp.MustCompile(`^[0-9]{12}$`)
	tahunAkademikRegex = regexp.MustCompile(`^[0-9]{4}/[0-9]{4}-(Ganjil|Genap)$`)
)

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func IsValidNIM(nim string) bool {
	return nimRegex.MatchString(nim)
}

func IsValidAngkatan(angkatan int) bool {
	currentYear := time.Now().Year()
	// Minimal angkatan yang masuk akal misal tahun 2000, maksimal tahun berjalan
	return angkatan >= 2000 && angkatan <= currentYear
}

func IsValidIPK(ipk float64) bool {
	return ipk >= 0.00 && ipk <= 4.00
}

func IsValidTahunAkademik(ta string) bool {
	if !tahunAkademikRegex.MatchString(ta) {
		return false
	}
	// Periksa kelogisan tahun (misal 2026/2027)
	year1, _ := strconv.Atoi(ta[0:4])
	year2, _ := strconv.Atoi(ta[5:9])
	return year2 == year1+1
}

// HitungBatasSKS menghitung batas SKS sesuai business rule:
// a. IPK terakhir: IPK >= 3,00 maksimal 24 SKS;
// b. IPK 2,50–2,99 maksimal 21 SKS;
// c. IPK < 2,50 maksimal 18 SKS.
func HitungBatasSKS(ipk float64) int {
	if ipk >= 3.00 {
		return 24
	} else if ipk >= 2.50 {
		return 21
	}
	return 18
}
