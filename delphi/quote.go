package delphi

const hex = "0123456789ABCDEF"

func Quote(s string) string {
	q := make([]byte, 0, len(s)+2)
	q = append(q, '\'')
	for _, b := range []byte(s) {
		if b < 0x20 || b > 0x7f || b == '\'' {
			q = append(q, '\'', '#', '$', hex[b>>4], hex[b&7], '\'')
		} else {
			q = append(q, b)
		}
	}
	q = append(q, '\'')
	return string(q)
}
