package main

import (
	"io"
	"net/http"
	"sync"
)

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


func asyncHttpGetCall(url string, ch chan<- Result) {

	resp, err := http.Get(url)
	if err != nil {
		ch <- Error[error]{Value: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- Error[error]{Value: err}
	}

	ch <- Ok[BodyStr]{Value: string(body)}

}

func AsyncChainOfHttpGetCalls(urls []string) []Result {

	results := make([]Result, len(urls))
	ch := make(chan Result, len(urls))

	for _, url := range urls {
		go asyncHttpGetCall(url, ch)
	}

	for i := 0; i < len(urls); i++ {
		results[i] = <-ch
	}

	close(ch)

	return results
}

func SyncChainOfHttpGetCalls(urls []string) []Result {
	var wg sync.WaitGroup
	results := make([]Result, len(urls))
	ch := make(chan Result, len(urls))

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			asyncHttpGetCall(url, ch)
		}(url)
	}

	wg.Wait()

	for i := 0; i < len(urls); i++ {
		results[i] = <-ch
	}

	close(ch)

	return results
}


func main(){
	
	// Examples of using the Result type
	urls := []string{
		"https://api.chucknorris.io/jokes/random",
		"https://api.chucknorris.io/jokes/random",
		"https://api.chucknorris.io/jokes/random",
	}
	
	// api calls
	var results []Result = SyncChainOfHttpGetCalls(urls)
	bodyResults := make([]BodyStr, 0)
	// Opera los resultados iterando sobre ellos
	for _, result := range results {
		switch result := result.(type) {
		case Ok[fn.BodyStr]:
			fmt.Println("Ok:", result.Value)
			// agrega el VALOR del resultado al slice bodyResults
			bodyResults = append(bodyResults, result.Value)
		case Error[string]:
			fmt.Println("Error:", result.Value)
		}
	}


}

/** 

WORK IN PROGRESS 

func SyncChainOfHttpGetCallsBodys(urls []string) Result {

	// api calls
	var results []Result = SyncChainOfHttpGetCalls(urls)

	unpackedResults := Map[Result](results, func(r Result) Result {
		switch r := r.(type) {
		case Ok[BodyStr]:
			return Ok[string]{Value: r.Value}
		case Error[string]:
			return Error[string]{Value: r.Value}
		}
		return Error[string]{Value: "Error desconocido"}
	})

	return Error[string]{Value: "Error desconocido"}
}

**/



