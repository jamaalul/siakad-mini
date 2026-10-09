package helper

// HasAnyRole memeriksa apakah role pengguna cocok dengan salah satu dari role yang diizinkan.
func HasAnyRole(role string, allowed ...string) bool {
	for _, a := range allowed {
		if role == a {
			return true
		}
	}
	return false
}
