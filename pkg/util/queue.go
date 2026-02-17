package util

type Queue chan string

func (q Queue) Enqueue(url string) {
	if url == "" {
		return
	}
	q <- url
}

func (q Queue) Dequeue() string {
	select {
	case url := <-q:
		return url
	default:
		return ""
	}
}
