package audit

import (
	"github.com/wptide/pkg/message"
	"io"
)

// Result is an interface map used to store results from processes.
type Result map[string]interface{}

// Processor is an interface for processing a message and producing results.
type Processor interface {
	Process(msg message.Message, result *Result)
	Kind() string
}

type PostProcessor interface {
	Processor
	SetReport(reader io.Reader)
	Parent(processor Processor)
}

func CanRunAudit(p Processor, results *Result) bool {
	r := *results
	if r["audits"] != nil {
		for _, v := range r["audits"].([]string) {
			if v == p.Kind() {
				return true
			}
		}
	}
	return false
}
