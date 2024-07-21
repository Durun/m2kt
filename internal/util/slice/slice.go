package slice

func Chunked[T any](s []T, size int) [][]T {
	if size <= 0 {
		return nil
	}

	chunks := make([][]T, 0, (len(s)+size-1)/size)
	for size < len(s) {
		s, chunks = s[size:], append(chunks, s[0:size:size])
	}
	if 0 < len(s) {
		chunks = append(chunks, s)
	}
	return chunks
}
