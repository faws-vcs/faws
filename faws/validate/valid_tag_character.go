package validate

func is_letter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func is_digit(r rune) bool {
	return (r >= '0' && r <= '9')
}

func is_invalid_tag_character(r rune) bool {
	return !(is_letter(r) || is_digit(r) || r == '.' || r == '_' || r == '-')
}
