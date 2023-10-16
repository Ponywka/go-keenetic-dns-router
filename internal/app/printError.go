package app

import (
	"errors"
	"fmt"
	"github.com/Ponywka/go-keenetic-dns-router/pkg/errors/contextedError"
	"github.com/Ponywka/go-keenetic-dns-router/pkg/errors/parentError"
)

func printError(err error) {
	var parentErr parentError.ParentError
	var contextedErr contextedError.ContextedError
	switch {
	case errors.As(err, &parentErr):
		fmt.Printf("%s: %s\r\n", parentErr.GetOrigin(), err.Error())
		printError(parentErr.GetChild())
	case errors.As(err, &contextedErr):
		fmt.Printf("%s: %s\r\n", contextedErr.GetOrigin(), err.Error())
	default:
		fmt.Println(err.Error())
	}
}
