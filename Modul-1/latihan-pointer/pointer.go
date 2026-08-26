package main

import "fmt"

func swapByValue(a, b int) {
	a, b = b, a
}
func swap(a, b *int) {
	*a, *b = *b, *a
}
func updateSliceByValue(s []string, newItem string) {
	s = append(s, newItem)
}
func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

func main() {
	fmt.Println("Demonstrasi swap")
	x, y := 10, 20

	fmt.Println("Data Awal             : x =", x, "| y =", y)
	
	swapByValue(x, y)
	fmt.Println("Setelah swapByValue   : x =", x, "| y =", y, "(Tidak berubah!)")
	
	swap(&x, &y)
	fmt.Println("Setelah swap (Pointer): x =", x, "| y =", y, "(Berhasil ditukar!)")

	fmt.Println("\nDemonstrasi updateSlice")
	techStack := []string{"Laravel", "PHP"}

	fmt.Printf("Data Awal                    : %v\n", techStack)
	
	updateSliceByValue(techStack, "Dart")
	fmt.Printf("Setelah updateSliceByValue   : %v (Dart tidak masuk!)\n", techStack)
	
	updateSlice(&techStack, "Flutter")
	fmt.Printf("Setelah updateSlice (Pointer): %v (Flutter berhasil masuk!)\n", techStack)
}