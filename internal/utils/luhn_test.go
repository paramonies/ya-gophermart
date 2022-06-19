package utils

import "testing"

func TestLuhn(t *testing.T) {
	validNumbers := []int{
		79927398713,
		4929972884676289,
		4532733309529845,
		4539088167512356,
		5577189519503182,
		5499078785968242,
		5236582963742210,
		379537021417898,
		373494930335082,
		379203612454689,
		6011223604226714,
		6011625707082028,
	}
	for _, number := range validNumbers {
		if !Valid(number) {
			t.Errorf("%v should be valid", number)
		}

		if CalculateLuhn(number/10) != number%10 {
			t.Errorf("%v's check number should be %v, but got %v", number, number%10, CalculateLuhn(number/10))
		}
	}
}
