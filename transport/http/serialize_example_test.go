package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/awslabs/smithy-go/middleware"
)

func ExampleRequest_serializeMiddleware() {
	// Create the stack and provide the function that will create a new Request
	// when the SerializeStep is invoked.
	stack := middleware.NewStack("serialize example", NewStackRequest)

	type Input struct {
		FooName  string
		BarCount int
	}

	// Add the serialization middleware.
	stack.Serialize.Add(middleware.SerializeMiddlewareFunc("example serialize",
		func(ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler) (
			middleware.SerializeOutput, error,
		) {
			req := in.Request.(*Request)
			input := in.Parameters.(*Input)

			req.Header.Set("foo-name", input.FooName)
			req.Header.Set("bar-count", strconv.Itoa(input.BarCount))

			return next.HandleSerialize(ctx, in)
		}),
		middleware.After,
	)

	// Mock example handler taking the request input and returning a response
	mockHandler := middleware.HandlerFunc(func(ctx context.Context, in interface{}) (
		output interface{}, err error,
	) {
		req := in.(*Request)
		fmt.Println("foo-name", req.Header.Get("foo-name"))
		fmt.Println("bar-count", req.Header.Get("bar-count"))

		return &Response{
			Response: &http.Response{
				StatusCode: 200,
				Header:     http.Header{},
			},
		}, nil
	})

	// Use the stack to decorate the handler then invoke the decorated handler
	// with the inputs.
	handler := middleware.DecorateHandler(mockHandler, stack)
	_, err := handler.Handle(context.Background(), &Input{FooName: "abc", BarCount: 123})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to call operation, %v", err)
		return
	}

	// Output:
	// foo-name abc
	// bar-count 123
}
