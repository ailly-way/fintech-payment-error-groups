package main

func main() {
	if err := runServer(); err != nil {
		panic(err)
	}
}
