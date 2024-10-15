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

type BodyStr = string

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


func UnpackResults(results []Result) ([]BodyStr, []error) {
	var bodyRequestResults []BodyStr
	var bodyRequestErrors []error

	for _, result := range results {
		switch result := result.(type) {
		case Ok[BodyStr]:
			bodyRequestResults = append(bodyRequestResults, result.Value)
			bodyRequestErrors = append(bodyRequestErrors, nil)
		case Error[error]:
			bodyRequestErrors = append(bodyRequestErrors, result.Value)
			bodyRequestResults = append(bodyRequestResults, "")
		}
	}

	return bodyRequestResults, bodyRequestErrors
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

	// Tour the result list and have anything with them (in this case it only prints)
	for _, result := range results {
		switch result := result.(type) {
		case Ok[BodyStr]:
			fmt.Println("Ok:", result.Value)
		case Error[error]:
			fmt.Println("Error:", result.Value)
		}
	}


	// or use the unpacking function
	uResults, uErrors := UnpackResults(results)


	// Print the results
	fmt.Println("Resultados desempaquetados:", uResults)
	fmt.Println("Errores desempaquetados:", uErrors)

	// and use the unpacked results in a traditional way
	for i := 0; i < len(uResults); i++ {
		if uErrors[i] != nil {
			fmt.Println("Error:", uErrors[i])
		} else {
			fmt.Println("Ok:", uResults[i])

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



