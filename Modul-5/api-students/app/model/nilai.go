package model

type Nilai struct {
	IDNilai        int     `json:"id_nilai"`
	NamaMataKuliah string  `json:"nama_mata_kuliah"`
	Nilai          float64 `json:"nilai"`
	IDStudent      int     `json:"id_student"`
}

type StudentNilaiResponse struct {
	Student     Student `json:"student"`
	DaftarNilai []Nilai `json:"daftar_nilai"`
}

type CreateNilaiRequest struct {
	NamaMataKuliah string  `json:"nama_mata_kuliah"`
	Nilai          float64 `json:"nilai"`
	IDStudent      int     `json:"id_student"`
}
