package kapbit

import (
	"testing"

	"github.com/kapbit/kapbit-go"
	wfl "github.com/kapbit/kapbit-go/workflow"
	"github.com/ymz-ncnk/mok"
)

type TestCase struct {
	Name   string
	Setup  Setup
	Params wfl.Params
	Want   Want
}

type Setup struct {
	Tools   kapbit.Tools
	Options kapbit.Options
	mock    []*mok.Mock
	Prepare func(k *kapbit.Kapbit)
}

type Want struct {
	Result wfl.Result
	Error  error
	Check  func(t *testing.T, k *kapbit.Kapbit)
}
