package queryresult

type ResultStreamer[T any] struct {
	Results            chan *Result[T]
	allResultsReceived chan string
}

func NewResultStreamer[T any]() *ResultStreamer[T] {
	return &ResultStreamer[T]{
		// make buffered channel so we can always stream a single result
		Results:            make(chan *Result[T], 1),
		allResultsReceived: make(chan string, 1),
	}
}

// StreamResult streams result on the Results channel, then waits for them to be read
func (q *ResultStreamer[T]) StreamResult(result *Result[T]) {
	q.Results <- result
	// wait for the result to be read
	<-q.allResultsReceived
}

// Close closes the result stream
func (q *ResultStreamer[T]) Close() {
	close(q.Results)
}

// AllResultsRead is a signal that indicates the all results have been read from the stream
func (q *ResultStreamer[T]) AllResultsRead() {
	q.allResultsReceived <- ""
}
