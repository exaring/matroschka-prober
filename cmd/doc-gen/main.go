package main

import (
	"fmt"
	"os"

	"github.com/projectdiscovery/yamldoc-go/encoder"
	"github.com/exaring/matroschka-prober/pkg/config"
)

func main() {
	FileDocs := []*encoder.FileDoc{
		config.GetconfigDoc(),
	}

	for _, fd := range FileDocs {
		fc, err := fd.Encode()
		if err != nil {
			fmt.Printf("failed to encode the file doc: %v", err)
		}

		err = os.WriteFile("docs/"+fd.Name+".md", fc, 0600)
		if err != nil {
			fmt.Printf("unable to write doc file: %v", err)
		}
	}
}
