package main

import (
	"io"
	"net/http"
	"sync"
)

/*

	The monadic Result interface is a data type that can be
	Ok or Error, it is a way to handle errors safely and controlled,
	without the need to use exceptions.

	This monada especially useful to implement gorutines and channels
	that are capable of avoiding panics and unexpected results.

	The Ok and Error types are data types that implement the
	Result interface, Ok is a data type that represents a correct value
	and Error is a data type that represents an error.

	This implementation is similar to the Maybe monad in Haskell and
	Result in Rust.

*/

type Result interface {
	isResult()
}
type Ok[T any] struct {
	Value T
}
type Error[U any] struct {
	Value U
}

func (Ok[T]) isResult()    {}
func (Error[U]) isResult() {}

/* ************************************** */

// Example of using the Result monad implemented in Go

// Interface to define the parameters of the AsyncHttpGetCall function
type UrlAndChanelParams interface {
	isUrlAndChanelParams()
}

// Structure that defines the parameters of the AsyncHttpGetCall function
type UrlAndChanel[T string, U chan<- Result] struct {
	Url T
	Ch  U
}

// Implementation of the UrlAndChanelParams interface
func (UrlAndChanel[T, U]) isUrlAndChanelParams() {}

// Alias for the RequestBodyAsString data type, which is a string
type RequestBodyAsString = string

// Asynchronous function that makes an HTTP GET request
// Using Goroutines and channels
// Receives a structure that contains the URL and a channel to send the result
// The function sends the result to the channel
// If an error occurs, it sends an error message to the channel
// The channel is closed at the end of the function
func AsyncHttpGetCall(params UrlAndChanelParams) {
	p := params.(UrlAndChanel[string, chan<- Result])
	url := p.Url
	ch := p.Ch
	resp, err := http.Get(url)
	if err != nil {
		ch <- Error[error]{Value: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- Error[error]{Value: err}
	}

	ch <- Ok[RequestBodyAsString]{Value: string(body)}
}

// Function that makes a chain of HTTP GET calls asynchronously
func AsyncChainOfHttpGetCalls(urls []string) []Result {
	results := make([]Result, len(urls))
	ch := make(chan Result, len(urls))
	for _, url := range urls {
		params := UrlAndChanel[string, chan<- Result]{Url: url, Ch: ch}
		go AsyncHttpGetCall(params)
	}
	for i := 0; i < len(urls); i++ {
		results[i] = <-ch
	}
	close(ch)
	return results
}

// Function that makes a chain of HTTP GET calls synchronously
// using the AsyncHttpGetCall function
// The function returns a slice of Result
// The function uses the UnpackResults function to get the results
// of the HTTP GET requests
func SyncChainOfHttpGetCalls(urls []string) []Result {
	var wg sync.WaitGroup
	results := make([]Result, len(urls))
	ch := make(chan Result, len(urls))
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			params := UrlAndChanel[string, chan<- Result]{Url: url, Ch: ch}
			AsyncHttpGetCall(params)
		}(url)
	}
	wg.Wait()
	for i := 0; i < len(urls); i++ {
		results[i] = <-ch
	}
	close(ch)
	return results
}

// Function that unpacks the results of the HTTP GET requests
// The function receives a slice of Result and returns two slices,
// one with the correct results and another with the errors
// The function uses the Ok and Error types to handle the results
func UnpackResults(results []Result) ([]RequestBodyAsString, []error) {
	var bodyRequestResults []RequestBodyAsString
	var bodyRequestErrors []error

	for _, result := range results {
		switch result := result.(type) {
		case Ok[RequestBodyAsString]:
			bodyRequestResults = append(bodyRequestResults, result.Value)
			bodyRequestErrors = append(bodyRequestErrors, nil)
		case Error[error]:
			bodyRequestErrors = append(bodyRequestErrors, result.Value)
			bodyRequestResults = append(bodyRequestResults, "")
		}
	}

	return bodyRequestResults, bodyRequestErrors
}

func main() {

	urls := []string{
		"https://api.chucknorris.io/jokes/random",
		"https://api.chucknorris.io/jokes/random",
		"https://api.chucknorris.io/jokes/random",
	}

	// Example of using the SyncChainOfHttpGetCalls function
	resultsSyncChainOfHttpGetCalls := SyncChainOfHttpGetCalls(urls)
	bodyRequestResults, bodyRequestErrors := UnpackResults(resultsSyncChainOfHttpGetCalls)
	for i, bodyRequestResult := range bodyRequestResults {
		if bodyRequestErrors[i] != nil {
			println("Error:", bodyRequestErrors[i])
		} else {
			println(bodyRequestResult)
		}
	}

	// Example of using the AsyncChainOfHttpGetCalls function
	// consider that this function can also use
	// UnpackResults if no additional processing is required
	resultsAsyncChainOfHttpGetCalls := AsyncChainOfHttpGetCalls(urls)
	for _, result := range resultsAsyncChainOfHttpGetCalls {
		switch result := result.(type) {
		case Ok[RequestBodyAsString]:
			println(result.Value)
		case Error[error]:
			println("Error:", result.Value)
		}
	}

	// Example of using the AsyncHttpGetCall function
	resultAsyncHttpGetCall := make(chan Result)
	params := UrlAndChanel[string, chan<- Result]{Url: "https://api.chucknorris.io/jokes/random", Ch: resultAsyncHttpGetCall}
	go AsyncHttpGetCall(params)
	result := <-resultAsyncHttpGetCall
	switch result := result.(type) {
	case Ok[RequestBodyAsString]:
		println(result.Value)
	case Error[error]:
		println("Error:", result.Value)
	}

}
