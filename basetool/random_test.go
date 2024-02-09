package basetool

import "testing"

func TestGenRand(t *testing.T) {
	t.Run("rand_string", func(t *testing.T) {
		length := 5
		rand1 := RandString(length)
		if len(rand1) != 5 {
			t.Error("length err")
			return
		}
		if RandString(length) == rand1 {
			t.Error("rand invalid")
			return
		}
	})

	t.Run("rand_letters", func(t *testing.T) {
		length := 5
		rand1 := RandomLetters(length)
		if len(rand1) != 5 {
			t.Error("length err")
			return
		}
		if RandomLetters(length) == rand1 {
			t.Error("rand invalid")
			return
		}
	})

	t.Run("rand_numbers", func(t *testing.T) {
		length := 5
		rand1 := RandomNumbers(length)
		if len(rand1) != 5 {
			t.Error("length err")
			return
		}
		if RandomNumbers(length) == rand1 {
			t.Error("rand invalid")
			return
		}
	})
}
