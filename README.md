# PlantUML Go Client
This project provides a handy CLI for PlantUML users. 

## Motivation 
* Self-contained tool
* Non-Java
* Able to work with hosted PlantUML server
* Produces "Text Format" 
* Produces Link
* Produced Images (Wow!)
* Learn [Go][http://golang.org] 

## How to use?

Get the go package first.

```shell
go get github.com/yogendra/plantuml-go
```

Now, run `plantuml-go`

```shell
plantuml-go my-uml.puml
```

or

```shell
echo "@startuml 
a -> b : hello world
@enduml" | plantuml-go
```

### How to generate images?

Use `-f png -o output` options on command line

```shell 
plantuml-go -f png -o output my-uml.puml
```

Above command will create a `my-uml.png` file next to `my-uml.puml` file (same directory).
 
## ToDo
* CD/CI
* Release
* Improve Test Coverage
