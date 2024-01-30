package main

func main() {
	app := NewApplication(nil)
	err := app.Start(":8080")
	if err != nil {
		return
	}
}
