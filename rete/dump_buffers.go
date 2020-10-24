package rete

import "fmt"


func DumpBuffers(root Node) {
	Walk(root, func(n Node) {
		if n, ok := n.(AbstractBufferNode); ok {
			fmt.Printf("node %s:\n", n.Label())
			n.DoItems(func(item interface{}) {
				fmt.Printf("    %s\n", item)
			})
		}
	})
}

