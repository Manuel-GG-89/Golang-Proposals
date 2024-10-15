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


func asyncHttpGetCall(url string, ch chan<- ty.Result) {

	resp, err := http.Get(url)
	if err != nil {
		ch <- ty.Error[error]{Value: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- ty.Error[error]{Value: err}
	}

	ch <- ty.Ok[BodyStr]{Value: string(body)}

}

func AsyncChainOfHttpGetCalls(urls []string) []ty.Result {

	results := make([]ty.Result, len(urls))
	ch := make(chan ty.Result, len(urls))

	for _, url := range urls {
		go asyncHttpGetCall(url, ch)
	}

	for i := 0; i < len(urls); i++ {
		results[i] = <-ch
	}

	close(ch)

	return results
}

func SyncChainOfHttpGetCalls(urls []string) []ty.Result {
	var wg sync.WaitGroup
	results := make([]ty.Result, len(urls))
	ch := make(chan ty.Result, len(urls))

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
		case ty.Ok[fn.BodyStr]:
			fmt.Println("Ok:", result.Value)
			// agrega el VALOR del resultado al slice bodyResults
			bodyResults = append(bodyResults, result.Value)
		case ty.Error[string]:
			fmt.Println("Error:", result.Value)
		}
	}


}

/** 

WORK IN PROGRESS 

func SyncChainOfHttpGetCallsBodys(urls []string) ty.Result {

	// api calls
	var results []ty.Result = SyncChainOfHttpGetCalls(urls)

	unpackedResults := Map[ty.Result](results, func(r ty.Result) ty.Result {
		switch r := r.(type) {
		case ty.Ok[BodyStr]:
			return ty.Ok[string]{Value: r.Value}
		case ty.Error[string]:
			return ty.Error[string]{Value: r.Value}
		}
		return ty.Error[string]{Value: "Error desconocido"}
	})

	return ty.Error[string]{Value: "Error desconocido"}
}

**/



