package main

import "fmt"

type Student struct {
	ID       int
	Name     string
	Grade    float64
	IsActive bool
}

func (s Student) GetInfo() string {
	status := "Non-Aktif"
	if s.IsActive {
		status = "Aktif"
	}
	return fmt.Sprintf("ID: %d | Nama: %s | Nilai: %.1f | Status: %s", s.ID, s.Name, s.Grade, status)
}

func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

func (s *Student) Activate() {
	s.IsActive = true
}

func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	mhs := Student{
		ID:       101,
		Name:     "Muhammad Fadhil Ilyas",
		Grade:    80.0,
		IsActive: false,
	}

	fmt.Println("Data Awal")
	fmt.Println(mhs.GetInfo())

	mhs.UpdateGrade(95.5)
	mhs.Activate()

	fmt.Println("\nSetelah Diperbarui")
	fmt.Println(mhs.GetInfo())

	mhs.Deactivate()

	fmt.Println("\n Setelah Dinonaktifkan (Deactivate)")
	fmt.Println(mhs.GetInfo())
}