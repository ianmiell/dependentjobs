package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	cgfile := "./examples/one-dep.yaml"
	if len(os.Args) > 1 {
		cgfile = os.Args[1]
	}
	fromfile(cgfile)

	// store the call graph:
	// err := dj.Store("./examples/dump.yaml")
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

func fromfile(cgfile string) {
	fmt.Printf("Creating call graph:\n")
	dj := New()
	err := dj.FromFile(cgfile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v\n", dj)
	fmt.Println("Running jobs in call graph:")
	dj.Run()
	dj.Complete()
	fmt.Printf("Call sequence: %v\n", dj.CallSeq())
}

func manualcg() DependentJobs {
	dj := New()
	dj.Add("root", "job 1", 0)
	dj.Add("j2", "job 2", 1)
	dj.Add("j3", "job 3", 1)
	dj.Add("j4", "job 4", 2)
	dj.Add("j5", "job 5", 2)
	dj.AddDependents("j4", "j5")
	dj.AddDependents("j2", "j4")
	dj.AddDependents("j3", "j4")
	dj.AddDependents("root", "j2", "j3", "j5")
	fmt.Printf("%#v\n", dj)
	return dj
}

func djcron() {
	// run the call graph and print the call sequence:
	jticks = make(map[string]int)
	cycle := 0
	for {
		go func() {
			fmt.Printf("\n--- CYCLE %d\nCreating call graph:\n", cycle)
			dj := New()
			// dj.Add("root", "job 1", 0)
			// dj.Add("j2", "job 2", 1)
			// dj.Add("j3", "job 3", 1)
			// dj.AddPeriodic("j2", 2)
			// dj.AddPeriodic("j3", 3)
			// dj.AddDependents("root", "j2")
			// dj.AddDependents("j2", "j3")
			err := dj.FromFile("./examples/simplecron.yaml")
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%#v\n", dj)
			fmt.Println("Running jobs in call graph:")
			dj.Run()
			dj.Complete()
			fmt.Printf("Call sequence: %v\n", dj.CallSeq())
		}()
		time.Sleep(1 * time.Second)
		cycle++
	}
}
