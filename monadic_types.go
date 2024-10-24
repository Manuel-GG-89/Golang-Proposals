package main

import (
	"fmt"
	"io"
	"net/http"
)

/*

   These types are just experimental, I have not tested their functionality
   and practical application, but they may be useful to someone, at least
   as a starting point for implementing monads in Go.

   It is necessary to mention that Go is not a purely functional language
   and it does not evaluate its functions lazily, so using monad types that
   depend on these concepts is not the most appropriate, but it can be done anyway.

   In this file, the IO and AccOperation monads are defined, as well as functions
   to map and reduce data collections.

*/

/*

   IO Monad

*/

// Monadic type IO, used to contextualize any
// input/output operation (safely, I think..)
type IO[A any] struct {
	run func() A
}

// Function to enter a value into the IO context
func Return[A any](value A) IO[A] {
	return IO[A]{run: func() A { return value }}
}

// Chain function belonging to the IO monad
// Used to chain input/output actions
// in Haskell it is called 'bind' and its operator is (>>=)
func (io IO[A]) Chain(f func(A) IO[A]) IO[A] {
	return IO[A]{run: func() A {
		return f(io.run()).run()
	}}
}

// Run function that executes the operation encapsulated
// within an IO context
func (io IO[A]) Run() A {
	return io.run()
}

/*
   Examples of IO Monad implementation
*/

// Example 1: Function to encapsulate an input operation
// that reads a line of text from the console
func ReadLine() IO[string] {
	return IO[string]{run: func() string {
		var input string
		fmt.Scanln(&input)
		return input
	}}
}

// Example 2: Function to encapsulate an output operation
// that prints a message to the console
func Println(message string) IO[string] {
	return IO[string]{run: func() string {
		fmt.Println(message)
		return message
	}}
}

/*

   AccOperation Monad

*/

// Monadic type AccOperation, used to chain
// operations that accumulate their result into a single final value
type AccOperation[T any] struct {
	accValue T
	err      error
}

// Function to create a new instance of AccOperation
// with an initial value and a possible error
func NewAccOperation[T any](accValue T, err error) AccOperation[T] {
	return AccOperation[T]{accValue: accValue, err: err}
}

// Function to chain accumulation operations
// in the AccOperation monad (similar to the Chain function of the IO monad)
// Receives a function that takes a value of type T and returns an AccOperation[T]
func (m AccOperation[T]) Chain(f func(T any) AccOperation[T]) AccOperation[T] {
	if m.err != nil {
		return AccOperation[T]{err: m.err}
	}
	return f(m.accValue)
}

// Function to execute the chained operations
// in the AccOperation monad and return the final accumulated value
func (m AccOperation[T]) Return() T {
	return m.accValue
}

/*
   Examples of AccOperation implementation
*/

// Asynchronous function that makes an HTTP GET request
// Using the AccOperation monad
func ChainedAsyncHttpGet(url string) AccOperation[string] {
	resp, err := http.Get(url)
	if err != nil {
		return NewAccOperation("", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewAccOperation("", err)
	}
	return NewAccOperation(string(body), nil)
}

/*

   Mappers and higher-order functions

*/

// Mapper is a function that takes an interface (a trait) that
// takes an input of any type and returns an output of any type.
type Mapper func(interface{}) interface{}

// MapAny applies a given Mapper function to each element of a slice of any type
// and returns a new slice with the results.
func MapAny(slice []interface{}, mapper Mapper) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = mapper(v)
	}
	return result
}

// Creates a Map function where by specifying a type, you can map
// a slice of that type and return a slice of the same type
func Map[T any](slice []T, mapper func(T) T) []T {
	result := make([]T, len(slice))
	for i, v := range slice {
		result[i] = mapper(v)
	}
	return result
}

// Creates a Reduce function that, taking a slice of a specific type,
// a reducer, and an initial value, can reduce the slice to a single value
func Reduce[T any, U any](slice []T, reducer func(U, T) U, initialValue U) U {
	result := initialValue
	for _, v := range slice {
		result = reducer(result, v)
	}
	return result
}

/* ************************************************************** */

// Structure that defines the parameters of the AsyncHttpGetCall function
type FunctionAndChanel[T func(), U chan<- Result] struct {
	Function T
	Ch       U
}

type AsyncIO[T FunctionAndChanel[func(), chan<- Result]] interface {
	Return(T) AsyncIO[T]
	Bind(func(T) AsyncIO[T]) AsyncIO[T]
	Map(func(T) T) AsyncIO[T]
}

// crea un tipo AsyncIOProcess que implementa la interfaz AsyncIO
type AsyncIOProcess[T FunctionAndChanel[func(), chan<- Result]] struct {
	value T
}

// función para retornar un valor en el contexto de AsyncIOProcess
func (a AsyncIOProcess[T]) Return(value T) AsyncIOProcess[T] {
	return NewAsyncIOProcess(value)
}

// función para encadenar operaciones en el contexto de AsyncIOProcess
func (a AsyncIOProcess[T]) Bind(f func(T) AsyncIOProcess[T]) AsyncIOProcess[T] {
	return NewAsyncIOProcess(f(a.value).value)
}

// función para crear una instancia de AsyncIOProcess
func NewAsyncIOProcess[T FunctionAndChanel[func(), chan<- Result]](value T) AsyncIOProcess[T] {
	return AsyncIOProcess[T]{value: value}
}

// función para mapear operaciones en el contexto de AsyncIOProcess
func (a AsyncIOProcess[T]) Map(f func(T) T) AsyncIOProcess[T] {
	return NewAsyncIOProcess(f(a.value))
}

func testing() {
	// Ejemplo de AsyncIOProcess
	// Se crea una instancia de AsyncIOProcess con una función y un canal
	// Se encadena una operación que recibe la función y el canal y los ejecuta
	// Se crea una instancia de AsyncIOProcess con una función y un canal
	// Se encadena una operación que recibe la función y el canal y los ejecuta
	NewAsyncIOProcess(FunctionAndChanel[func(), chan<- Result]{
		Function: func() {
			fmt.Println("Hello, world!")
		},
		Ch: make(chan Result),
	}).Bind(func(f FunctionAndChanel[func(), chan<- Result]) AsyncIOProcess[FunctionAndChanel[func(), chan<- Result]] {
		f.Function()
		return NewAsyncIOProcess(f)
	}).Bind(func(f FunctionAndChanel[func(), chan<- Result]) AsyncIOProcess[FunctionAndChanel[func(), chan<- Result]] {
		f.Function()
		return NewAsyncIOProcess(f)
	}).Bind(func(f FunctionAndChanel[func(), chan<- Result]) AsyncIOProcess[FunctionAndChanel[func(), chan<- Result]] {
		f.Function()
		return NewAsyncIOProcess(f)
	}).Bind(func(f FunctionAndChanel[func(), chan<- Result]) AsyncIOProcess[FunctionAndChanel[func(), chan<- Result]] {
		f.Function()
		return NewAsyncIOProcess(f)
	} /* ... */).value.Function()

}
