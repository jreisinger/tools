package tools

// adapted from chatgpt on 2023-10-24
func Reverse(s string) string {
	runes := []rune(s)

	length := len(runes)
	reversed := make([]rune, length)

	for i, r := range runes {
		reversed[length-1-i] = r
	}

	return string(reversed)
}

// https://stackoverflow.com/questions/1752414/how-to-reverse-a-string-in-go
func Reverse2(s string) string {
	runes := []rune(s)

	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}
